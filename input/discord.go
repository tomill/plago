package input

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/dghubble/sling"
	"github.com/tomill/centre/config"
	"github.com/tomill/centre/message"
)

type Discord struct {
	since  time.Time
	until  time.Time
	tz     *time.Location
	client *sling.Sling
	sheet  *config.Sheet
}

func DiscordFetcher(c config.Config) (Fetcher, error) {
	p := &Discord{
		since: c.Since,
		until: c.Until,
		tz:    c.TimeZone,
		client: sling.New().
			Base("https://discord.com/").
			Set("Authorization", c.DiscordToken).
			Set("UserAgent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 Safari/537.36"),
	}

	// TODO move to main
	sheet, err := config.NewSheet(c.SheetServiceAccountKey, c.SheetID)
	if err != nil {
		return nil, err
	}
	p.sheet = sheet

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
		UserName string `json:"username"`
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

func (p Discord) Fetch() (message.Timeline, error) {
	timeline := message.Timeline{
		Source:  "discord",
		Subject: p.since.Format(time.DateOnly),
	}

	log.Println("get channel setting from sheet ...")
	channels, err := p.sheet.GetLists("discord.channels", DiscordChannel{})
	if err != nil {
		return timeline, err
	}

	for _, v := range channels {
		ch := v.(DiscordChannel)

		var messages []DiscordMessage
		if res, err := p.client.New().Get("api/v9/channels/" + ch.ChannelID + "/messages?limit=50").ReceiveSuccess(&messages); err != nil {
			return timeline, fmt.Errorf("discord web api call error: %w", err)
		} else if res.StatusCode != http.StatusOK {
			return timeline, fmt.Errorf("request error: %s - %s", res.Request.URL.Path, res.Status)
		}

		for _, m := range messages {
			if m.Timestamp.Before(p.since) || m.Timestamp.After(p.until) {
				continue
			}

			msg := message.Message{
				Section:   fmt.Sprintf("%s #%s", ch.ServerName, ch.ChannelName),
				Timestamp: m.Timestamp,
				Permalink: fmt.Sprintf("https://discord.com/channels/%s/%s/%s", ch.ServerID, m.ChannelID, m.ID),
				Lead:      m.Timestamp.In(p.tz).Format("15:04"),
				UserName:  m.Author.UserName,
				Text:      m.Content,
				Reply:     m.Reference.MessageID != "",
			}

			msg.Text = emoji.ReplaceAllString(msg.Text, `$1`)

			for _, v := range m.Mentions {
				msg.Text = strings.ReplaceAll(msg.Text, fmt.Sprintf("<@%s>", v.ID), "@"+v.UserName)
			}

			for _, v := range m.Attachments {
				if strings.HasPrefix(v.ContentType, "image/") {
					msg.Attachments = append(msg.Attachments, message.Message{
						Type:      message.TypeImage,
						Permalink: v.URL,
					})
				} else {
					msg.Attachments = append(msg.Attachments, message.Message{
						Text: v.Filename,
					})
				}
			}

			timeline.Append(msg)
		}
	}

	return timeline.Sorted(), nil
}

var (
	emoji = regexp.MustCompile(`<(:[^:]+:)\d+>`)
)

func (p Discord) build(ch DiscordChannel, post DiscordMessage) *message.Message {
	if post.Timestamp.Before(p.since) || post.Timestamp.After(p.until) {
		return nil
	}

	msg := &message.Message{
		Section:   fmt.Sprintf("%s #%s", ch.ServerName, ch.ChannelName),
		Timestamp: post.Timestamp,
		Permalink: fmt.Sprintf("https://discord.com/channels/%s/%s/%s", ch.ServerID, post.ChannelID, post.ID),
		UserName:  post.Author.UserName,
		Text:      post.Content,
		Reply:     post.Reference.MessageID != "",
	}

	msg.Text = emoji.ReplaceAllString(msg.Text, "$1")

	for _, v := range post.Mentions {
		msg.Text = strings.ReplaceAll(msg.Text, fmt.Sprintf("<@%s>", v.ID), "@"+v.UserName)
	}

	for _, v := range post.Attachments {
		if strings.HasPrefix(v.ContentType, "image/") {
			msg.Attachments = append(msg.Attachments, message.Message{
				Type:      message.TypeImage,
				Permalink: v.URL,
			})
		} else {
			msg.Attachments = append(msg.Attachments, message.Message{
				Text: v.Filename,
			})
		}
	}

	return msg
}
