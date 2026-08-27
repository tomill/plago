package input

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/samber/lo"
	"github.com/slack-go/slack"
	"github.com/tomill/plago"
	"github.com/tomill/plago/internal/config"
)

type Slack struct {
	config.ExecParams
	workspace  string
	client     *slack.Client
	channelIDs []string
	users      map[string]string
}

func SlackFetcher(c config.Config) (Fetcher, error) {
	p := &Slack{
		ExecParams: c.ExecParams,
		workspace:  c.SlackWorkspace,
		client:     slack.New(c.SlackToken, slack.OptionHTTPClient(httpClient)),
		channelIDs: c.SlackChannelIDs,
	}
	if p.channelIDs == nil {
		for _, ch := range c.SlackChannels {
			p.channelIDs = append(p.channelIDs, ch.ChannelID)
		}
	}

	return p, nil
}

func (p Slack) Fetch() (plago.Timeline, error) {
	timeline := newTimeline(p.ExecParams)

	p.users = lo.Must(p.getUsers())

	for _, channelID := range p.channelIDs {
		ch := p.channel(channelID)

		log.Printf("slack GetConversationHistory() %s %s ...", ch.ChannelID, ch.ChannelName)
		res, err := p.client.GetConversationHistory(&slack.GetConversationHistoryParameters{
			ChannelID: channelID,
			Oldest:    fmt.Sprintf("%d.0", p.Since.Unix()),
			Latest:    fmt.Sprintf("%d.0", p.Until.Unix()),
			Limit:     1000,
		})
		if err != nil {
			timeline.AppendError(err)
			continue
		}
		for _, post := range res.Messages {
			timeline.Append(p.build(ch, post))
		}

		time.Sleep(1 * time.Second)
	}

	return timeline.Sorted(), nil
}

func (p Slack) build(ch SlackChannel, post slack.Message) *plago.Entry {
	ts := p.time(post.Timestamp)

	entry := &plago.Entry{
		Section:   ts.Format("2006-01-02 15:00"),
		Channel:   ch.ChannelName,
		Timestamp: ts,
		URL:       fmt.Sprintf("https://%s.slack.com/archives/%s/p%s", p.workspace, post.Channel, strings.ReplaceAll(post.Timestamp, ".", "")),
		Reply:     post.ThreadTimestamp != "",
		User:      lo.CoalesceOrEmpty(p.users[post.User], post.Username),
	}

	if post.Text == "このメッセージにはインタラクティブ要素が含まれます。" {
		if blocks := post.Blocks.BlockSet; len(blocks) > 0 {
			if _, ok := blocks[0].(*slack.RichTextBlock); ok {
				if _, ok := blocks[0].(*slack.RichTextBlock).Elements[0].(*slack.RichTextSection); ok {
					for _, elem := range blocks[0].(*slack.RichTextBlock).Elements[0].(*slack.RichTextSection).Elements {
						switch elem := elem.(type) {
						case *slack.RichTextSectionTextElement:
							entry.Text = elem.Text
						case *slack.RichTextSectionUserElement:
							entry.Text = "@" + p.users[elem.UserID]
						}
					}
				}
			}
		}
	} else {
		entry.Text = p.text(post.Text)
	}

	for _, v := range post.Reactions {
		entry.Text += fmt.Sprintf(" [%s]", v.Name)
	}

	for _, v := range post.Attachments {
		if v.Text != "" {
			if v.Pretext != "" {
				entry.Text += " " + p.text(v.Pretext)
			}
			entry.AddAttachment(plago.Entry{Text: p.text(v.Title) + "\n" + p.text(v.Text)})
		} else if v.Fallback != "" {
			entry.AddAttachment(plago.Entry{Text: p.text(v.Fallback)})
		}
	}

	if entry.Text == "" && len(post.Attachments) == 0 {
		return nil
	}

	return entry
}

func (p Slack) time(ts string) time.Time {
	i, _ := strconv.ParseFloat(ts, 64)
	return time.Unix(int64(i), 0).In(tz)
}

type SlackChannel struct {
	ChannelID   string
	ChannelName string
}

func (p Slack) channel(cid string) SlackChannel {
	ch := SlackChannel{ChannelID: cid}
	channel, err := p.client.GetConversationInfo(&slack.GetConversationInfoInput{
		ChannelID:     cid,
		IncludeLocale: false,
	})
	if err != nil {
		ch.ChannelName = fmt.Sprintf("<%s>", cid)
		return ch
	}

	prefix := "#"
	switch {
	case channel.IsPrivate:
		prefix = "$"
	case channel.IsGroup:
		prefix = "%"
	}

	ch.ChannelName = prefix + channel.Name
	return ch
}

var (
	reSlackText   = regexp.MustCompile(`<[^>]+>`)
	reGithubImage = regexp.MustCompile(`(https://private-user-images.githubusercontent.com/[^?]+\?jwt=)\S+`)
)

func (p Slack) text(s string) string {
	s = strings.ReplaceAll(s, "```", "")
	s = reGithubImage.ReplaceAllString(s, `$1`)

	return reSlackText.ReplaceAllStringFunc(s, func(s string) string {
		s = s[1 : len(s)-1]
		ss := strings.SplitN(s, "|", 2)

		if strings.HasPrefix(ss[0], "@") {
			if u := p.users[ss[0][1:]]; u != "" {
				return "@" + u
			} else {
				return ss[0]
			}
		}
		if strings.HasPrefix(ss[0], "#") {
			return fmt.Sprintf("<#%s>", ss[0][1:])
		}
		if strings.HasPrefix(ss[0], "http://") || strings.HasPrefix(ss[0], "https://") {
			return ss[0]
		}
		if strings.HasPrefix(s, "!subteam") && len(ss) > 1 {
			return ss[1]
		}
		if strings.HasPrefix(s, "!") {
			return "@" + ss[0][1:]
		}

		return s
	})
}

func (p Slack) getUsers() (map[string]string, error) {
	users := map[string]string{}
	defer func() {
		log.Printf("%d vuser(s)", len(users))
	}()

	log.Printf("slack GetUsers() ...")
	res, err := p.client.GetUsers()
	if err != nil {
		return nil, err
	}
	for _, user := range res {
		if user.IsBot {
			continue
		}
		users[user.ID] = user.Name
	}

	return users, nil
}
