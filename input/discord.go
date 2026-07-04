package input

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/dghubble/sling"

	"github.com/tomill/centre/config"
	"github.com/tomill/centre/message"
)

type Discord struct {
	since    time.Time
	until    time.Time
	client   *sling.Sling
	channels []DiscordChannel
	users    map[string]string
}

func DiscordFetcher(c config.Config) (Fetcher, error) {
	p := &Discord{
		since: c.Since,
		until: c.Until,
		users: map[string]string{},
		client: sling.New().
			Base("https://discord.com/").
			Set("Authorization", c.DiscordToken).
			Set("UserAgent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36"),
	}

	if os.Getenv("DEBUG") == "1" {
		p.client = p.client.Client(newDebugClient())
	}

	res, err := c.Sheet().GetLists("discord.channels", DiscordChannel{})
	if err != nil {
		return nil, err
	}
	for _, v := range res {
		p.channels = append(p.channels, v.(DiscordChannel))
	}

	return p, nil
}

type DiscordChannel struct {
	ServerID    string
	ServerName  string
	ChannelID   string
	ChannelName string
	Fetch       string
}

type DiscordMessage struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	ChannelID string    `json:"channel_id"`
	Content   string    `json:"content"`
	Author    struct {
		ID         string `json:"id"`
		UserName   string `json:"username"`
		GlobalName string `json:"global_name"`
	} `json:"author"`
	Attachments []struct {
		Filename    string `json:"filename"`
		URL         string `json:"url"`
		ProxyURL    string `json:"proxy_url"`
		ContentType string `json:"content_type"`
	} `json:"attachments"`
	Embeds []struct {
		Type        string `json:"type"`
		URL         string `json:"url,omitempty"`
		Title       string `json:"title,omitempty"`
		Description string `json:"description,omitempty"`
		Image       *struct {
			ProxyURL    string `json:"proxy_url,omitempty"`
			ContentType string `json:"content_type,omitempty"`
		} `json:"image,omitempty"`
		Thumbnail *struct {
			ProxyURL    string `json:"proxy_url,omitempty"`
			ContentType string `json:"content_type,omitempty"`
		} `json:"thumbnail,omitempty"`
	} `json:"embeds"`
	Mentions []struct {
		ID       string `json:"id"`
		UserName string `json:"username"`
	} `json:"mentions"`
	StickerItems []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"sticker_items"`
	Reference struct {
		ChannelID string `json:"channel_id"`
		GuildID   string `json:"guild_id"`
		MessageID string `json:"message_id"`
	} `json:"message_reference"`
}

type DiscordMember struct {
	Nick string `json:"nick"`
	User struct {
		ID       string `json:"id"`
		UserName string `json:"username"`
	} `json:"user"`
}

func (p Discord) Fetch() (message.Timeline, error) {
	timeline := message.Timeline{
		Source:  "discord",
		Subject: p.since.Format(time.DateOnly),
	}

	for _, ch := range p.channels {
		var messages []DiscordMessage
		if res, err := p.client.New().Get("api/v10/channels/" + ch.ChannelID + "/messages?limit=50").ReceiveSuccess(&messages); err != nil {
			return timeline, fmt.Errorf("discord web api call error: %w", err)
		} else if res.StatusCode != http.StatusOK {
			return timeline, fmt.Errorf("request error: %s - %s", res.Request.URL.Path, res.Status)
		}

		for _, v := range messages {
			if msg := p.build(ch, v); msg != nil {
				timeline.Append(*msg)
			}
		}
	}

	return timeline.Sorted(), nil
}

var (
	emoji          = regexp.MustCompile(`<(:[^:]+:)\d+>`)
	markdownEscape = regexp.MustCompile(`\\([_*\[\]()~` + "`" + `>#+\-=|{}.!])`)
)

func (p Discord) build(ch DiscordChannel, post DiscordMessage) *message.Message {
	if post.Timestamp.Before(p.since) || post.Timestamp.Equal(p.until) || post.Timestamp.After(p.until) {
		return nil
	}

	msg := &message.Message{
		Timestamp: post.Timestamp.In(tz),
		Section:   post.Timestamp.In(tz).Format("2006-01-02 15:00"),
		Channel:   ch.ServerName,
		Permalink: fmt.Sprintf("https://discord.com/channels/%s/%s/%s", ch.ServerID, post.ChannelID, post.ID),
		UserName:  p.user(ch.ServerID, post.Author.ID, post.Author.GlobalName, post.Author.UserName),
		Text:      post.Content,
		Reply:     post.Reference.MessageID != "",
	}

	msg.Text = emoji.ReplaceAllString(msg.Text, `$1`)

	for _, v := range post.Mentions {
		msg.Text = strings.ReplaceAll(msg.Text, fmt.Sprintf("<@%s>", v.ID), "@"+v.UserName)
	}

	msg.Text = "[" + ch.ChannelName + "] " + msg.Text

	for _, v := range post.Attachments {
		if strings.HasPrefix(v.ContentType, "image/") {
			msg.AddAttachment(message.Message{
				Type:      message.TypeImage,
				Permalink: v.ProxyURL,
			})
		} else {
			msg.AddAttachment(message.Message{
				Text: v.Filename,
			})
		}
	}

	for _, v := range post.Embeds {
		msg.AddAttachment(message.Message{
			Text: strings.Join([]string{v.Title, markdownEscape.ReplaceAllString(v.Description, "$1")}, "\n"),
		})

		if v.Image != nil && strings.HasPrefix(v.Image.ContentType, "image/") {
			msg.AddAttachment(message.Message{
				Type:      message.TypeImage,
				Permalink: v.Image.ProxyURL,
			})
		} else if v.Thumbnail != nil && strings.HasPrefix(v.Thumbnail.ContentType, "image/") {
			msg.AddAttachment(message.Message{
				Type:      message.TypeImage,
				Permalink: v.Thumbnail.ProxyURL,
			})
		}
	}

	for _, v := range post.StickerItems {
		msg.AddAttachment(message.Message{
			Type:      message.TypeImage,
			Permalink: fmt.Sprintf(`https://media.discordapp.net/stickers/%s.png?size=320&passthrough=false`, v.ID),
		})
	}

	return msg
}

func (p Discord) user(gid, uid, global, username string) string {
	if nick, ok := p.users[uid]; ok {
		return nick
	}

	var member DiscordMember
	req := p.client.New().Get("api/v10/guilds/" + gid + "/members/" + uid)

	fallback := global
	if fallback == "" {
		fallback = username
	}

	if _, err := req.ReceiveSuccess(&member); err != nil {
		return fallback
	}

	if member.Nick != "" {
		p.users[uid] = member.Nick
	} else {
		p.users[uid] = fallback
	}

	return p.users[uid]
}
