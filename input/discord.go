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
	channels []DiscordChannel
	users    *memomap[string]
}

type DiscordChannel struct {
	ServerID    string
	ServerName  string
	ChannelID   string
	ChannelName string
}

func DiscordFetcher(c config.Config) (Fetcher, error) {
	p := &Discord{
		ExecParams: c.ExecParams,
		client:     lo.Must(discordgo.New(c.DiscordToken)),
		channels:   config.MustGetSheetValues[DiscordChannel](c.DiscordChannelSheet),
		users:      newMemomap[string](),
	}

	return p, nil
}

func (p Discord) Fetch() (entry.Timeline, error) {
	timeline := entry.NewTimeline(p.ExecParams)

	for _, ch := range p.channels {
		messages, err := p.client.ChannelMessages(ch.ChannelID, 100, "", "", "")
		if err != nil {
			return timeline, err
		}
		for _, v := range messages {
			timeline.Append(p.build(ch, v))
		}
	}

	return timeline.Sorted(), nil
}

var (
	reEmoji = regexp.MustCompile(`<(:[^:]+:)\d+>`)
)

func (p Discord) build(ch DiscordChannel, post *discordgo.Message) *entry.Entry {
	ts := post.Timestamp.In(tz)
	if !timeinrange(ts, p.ExecParams) {
		return nil
	}

	e := &entry.Entry{
		Timestamp: ts,
		Section:   ts.Format("2006-01-02 15:00"),
		Channel:   fmt.Sprintf("[%s] %s", ch.ServerName, ch.ChannelName),
		URL:       fmt.Sprintf("https://discord.com/channels/%s/%s/%s", ch.ServerID, post.ChannelID, post.ID),
		Reply:     post.ReferencedMessage != nil,
		UserName: p.users.get(ch.ServerID+post.Author.ID, func() string {
			member, err := p.client.GuildMember(ch.ServerID, post.Author.ID)
			if err == nil && member.Nick != "" {
				return member.Nick
			}
			return post.Author.DisplayName()
		}),
		Text: pipe(post.ContentWithMentionsReplaced(),
			func(s string) string {
				return reEmoji.ReplaceAllString(s, `$1`)
			},
		),
	}

	for _, v := range post.Attachments {
		if strings.HasPrefix(v.ContentType, "image/") {
			e.AddImage(v.ProxyURL)
		} else {
			e.AddAttachment(entry.Entry{
				Text: v.Filename,
			})
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
		e.AddAttachment(entry.Entry{
			Text: strings.Join([]string{v.Author.Name, v.Title, markdownUnescape(v.Description)}, "\n"),
		})

		if v.Image != nil {
			e.AddImage(v.Image.ProxyURL)
		} else if v.Thumbnail != nil {
			e.AddImage(v.Thumbnail.ProxyURL)
		}
	}

	return e
}
