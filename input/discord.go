package input

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/tomill/centre/config"
	"github.com/tomill/centre/message"
)

type Discord struct {
	since    time.Time
	until    time.Time
	session  *discordgo.Session
	channels []DiscordChannel
	users    map[string]string
}

type DiscordChannel struct {
	ServerID    string
	ServerName  string
	ChannelID   string
	ChannelName string
}

func DiscordFetcher(c config.Config) (Fetcher, error) {
	session, err := discordgo.New(c.DiscordToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create discord session: %w", err)
	}

	p := &Discord{
		since:   c.Since,
		until:   c.Until,
		users:   map[string]string{},
		session: session,
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

func (p Discord) Fetch() (message.Timeline, error) {
	timeline := message.Timeline{
		Source:  "discord",
		Subject: p.since.Format(time.DateOnly),
	}

	for _, ch := range p.channels {
		channel, err := p.session.Channel(ch.ChannelID)
		if err != nil {
			continue
		}
		if ch.ChannelName == "" {
			ch.ChannelName = channel.Name
		}

		if ch.ServerName == "" {
			guild, err := p.session.Guild(ch.ServerID)
			if err != nil {
				ch.ServerName = ch.ServerID
			}
			ch.ServerName = guild.Name
		}

		messages, err := p.session.ChannelMessages(ch.ChannelID, 50, "", "", "")
		if err != nil {
			return timeline, fmt.Errorf("discord api call error: %w", err)
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

func (p Discord) build(ch DiscordChannel, post *discordgo.Message) *message.Message {
	ts := post.Timestamp.In(tz)
	if ts.Before(p.since) || ts.Equal(p.until) || ts.After(p.until) {
		return nil
	}

	msg := &message.Message{
		Timestamp: ts,
		Section:   ts.Format("2006-01-02 15:00"),
		Channel:   ch.ServerName,
		Permalink: fmt.Sprintf("https://discord.com/channels/%s/%s/%s", ch.ServerID, post.ChannelID, post.ID),
		UserName:  p.user(ch.ServerID, post.Author),
		Text:      post.ContentWithMentionsReplaced(),
		Reply:     post.ReferencedMessage != nil,
	}

	msg.Text = emoji.ReplaceAllString(msg.Text, `$1`)

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

		if v.Image != nil {
			msg.AddAttachment(message.Message{
				Type:      message.TypeImage,
				Permalink: v.Image.ProxyURL,
			})
		} else if v.Thumbnail != nil {
			msg.AddAttachment(message.Message{
				Type:      message.TypeImage,
				Permalink: v.Thumbnail.ProxyURL,
			})
		}
	}

	for _, v := range post.StickerItems {
		switch v.FormatType {
		case discordgo.StickerFormatTypeGIF:
			msg.AddAttachment(message.Message{
				Type:      message.TypeImage,
				Permalink: fmt.Sprintf(`https://media.discordapp.net/stickers/%s.gif?size=320&passthrough=false`, v.ID),
			})
		default:
			msg.AddAttachment(message.Message{
				Type:      message.TypeImage,
				Permalink: fmt.Sprintf(`https://cdn.discordapp.net/stickers/%s.png?size=320&passthrough=false`, v.ID),
			})
		}
	}

	return msg
}

func (p Discord) user(gid string, user *discordgo.User) string {
	if nick, ok := p.users[user.ID]; ok {
		return nick
	}

	member, err := p.session.GuildMember(gid, user.ID)
	if err != nil || member.Nick == "" {
		p.users[user.ID] = user.DisplayName()
	} else {
		p.users[user.ID] = member.Nick
	}

	return p.users[user.ID]
}
