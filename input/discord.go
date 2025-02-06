package input

import (
	"fmt"
	"net/http"
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
			Set("UserAgent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 Safari/537.36"),
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
		ContentType string `json:"content_type"`
	} `json:"attachments"`
	Mentions []struct {
		ID       string `json:"id"`
		UserName string `json:"username"`
	} `json:"mentions"`
	Reference struct {
		ChannelID string `json:"channel_id"`
		GuildID   string `json:"guild_id"`
		MessageID string `json:"message_id"`
	} `json:"message_reference"`
}

type DiscordUser struct {
	GuildMember struct {
		Nick string `json:"nick"`
	} `json:"guild_member"`
}

func (p Discord) Fetch() (message.Timeline, error) {
	timeline := message.Timeline{
		Source:  "discord",
		Subject: p.since.Format(time.DateOnly),
	}

	for _, ch := range p.channels {
		var messages []DiscordMessage
		if res, err := p.client.New().Get("api/v9/channels/" + ch.ChannelID + "/messages?limit=50").ReceiveSuccess(&messages); err != nil {
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
	emoji = regexp.MustCompile(`<(:[^:]+:)\d+>`)
)

func (p Discord) build(ch DiscordChannel, post DiscordMessage) *message.Message {
	if post.Timestamp.Before(p.since) || post.Timestamp.Equal(p.until) || post.Timestamp.After(p.until) {
		return nil
	}

	msg := &message.Message{
		Timestamp: post.Timestamp,
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

	msg.Text += " #" + ch.ChannelName

	for _, v := range post.Attachments {
		if strings.HasPrefix(v.ContentType, "image/") {
			msg.AddAttachment(message.Message{
				Type:      message.TypeImage,
				Permalink: v.URL,
			})
		} else {
			msg.AddAttachment(message.Message{
				Text: v.Filename,
			})
		}
	}

	return msg
}

func (p Discord) user(gid, uid, global, username string) string {
	if nick, ok := p.users[uid]; ok {
		return nick
	}

	query := struct {
		Mutual  bool   `url:"with_mutual_guilds"`
		Count   bool   `url:"with_mutual_friends_count"`
		GuildID string `url:"guild_id"`
	}{
		Mutual:  true,
		Count:   false,
		GuildID: gid,
	}

	var user DiscordUser
	req := p.client.New().Get("api/v9/users/" + uid + "/profile").QueryStruct(query)

	fallback := global
	if fallback == "" {
		fallback = username
	}

	if _, err := req.ReceiveSuccess(&user); err != nil {
		return fallback
	}

	if user.GuildMember.Nick != "" {
		p.users[uid] = user.GuildMember.Nick
	} else {
		p.users[uid] = fallback
	}

	return p.users[uid]
}
