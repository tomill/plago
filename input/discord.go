package input

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/samber/lo"
	"github.com/tomill/centre/config"
	"github.com/tomill/centre/entry"
)

type Discord struct {
	config.ExecParams
	client   *discordgo.Session
	target   config.List
	channels *cache[DiscordChannel]
	users    *cache[string]
}

type DiscordChannel struct {
	ChannelID   string
	ChannelName string
	ServerID    string
	ServerName  string
}

func DiscordFetcher(c config.Config) (Fetcher, error) {
	p := &Discord{
		ExecParams: c.ExecParams,
		client:     lo.Must(discordgo.New(c.DiscordToken)),
		target:     c.DiscordChannels,
		channels:   newCache[DiscordChannel](),
		users:      newCache[string](),
	}

	p.client.Client = httpClient
	return p, nil
}

func (p Discord) Fetch() (entry.Timeline, error) {
	timeline := entry.NewTimeline(p.ExecParams)

	for _, channelID := range p.target {
		messages, err := p.client.ChannelMessages(channelID, 100, "", "", "")
		if err != nil {
			timeline.Append(entry.Error(err))
			continue
		}
		for _, v := range messages {
			timeline.Append(p.build(v))
		}
	}

	return timeline.Sorted(), nil
}

func (p Discord) build(post *discordgo.Message) *entry.Entry {
	ts := post.Timestamp.In(tz)
	if !timeinrange(ts, p.ExecParams) {
		return nil
	}

	ch := p.channel(post.ChannelID)

	e := &entry.Entry{
		Section:   ts.Format("2006-01-02 15:00"),
		Channel:   fmt.Sprintf("[%s] %s", ch.ServerName, ch.ChannelName),
		Timestamp: ts,
		URL:       fmt.Sprintf("https://discord.com/channels/%s/%s/%s", ch.ServerID, ch.ChannelID, post.ID),
		Reply:     post.ReferencedMessage != nil,
		User:      p.name(ch.ServerID, post.Author.ID, post.Author.DisplayName()),
		Text:      p.text(post.ContentWithMentionsReplaced()),
	}

	for _, v := range post.Attachments {
		if strings.HasPrefix(v.ContentType, "image/") {
			e.AddImage(v.ProxyURL)
		} else {
			e.AddAttachment(entry.Entry{Text: v.URL + "\n" + v.Filename})
		}
	}

	for _, v := range post.StickerItems {
		switch v.FormatType {
		case discordgo.StickerFormatTypeGIF:
			e.AddImage(fmt.Sprintf(`https://media.discordapp.net/stickers/%s.gif?size=320&passthrough=false`, v.ID))
		default:
			e.AddImage(fmt.Sprintf(`https://cdn.discordapp.net/stickers/%s.png?size=320&passthrough=false`, v.ID))
		}
	}

	for _, v := range post.Embeds {
		a := e.AddAttachment(entry.Entry{
			Text: strings.Join([]string{v.Title, markdownUnescape(v.Description)}, "\n"),
		})
		if v.Thumbnail != nil {
			a.AddImage(v.Thumbnail.ProxyURL)
		}
	}

	return e
}

func (p Discord) channel(cid string) DiscordChannel {
	return p.channels.get(cid, func() DiscordChannel {
		ch := DiscordChannel{ChannelID: cid}

		channel, err := p.client.Channel(cid)
		if err != nil {
			return ch
		}

		ch.ChannelName = channel.Name
		ch.ServerID = channel.GuildID

		server, err := p.client.Guild(channel.GuildID)
		if err != nil {
			return ch
		}

		ch.ServerName = server.Name
		return ch
	})
}

func (p Discord) name(sid, uid, fallback string) string {
	return p.users.get(uid, func() string {
		member, err := p.client.GuildMember(sid, uid)
		if err == nil && member.Nick != "" {
			return member.Nick
		}
		return fallback
	})
}

var (
	reEmoji = regexp.MustCompile(`<(:[^:]+:)\d+>`)
)

func (p Discord) text(s string) string {
	return reEmoji.ReplaceAllString(s, `$1`)
}
