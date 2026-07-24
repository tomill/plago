package input

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/dghubble/oauth1"
	"github.com/g8rswimmer/go-twitter/v2"
	"github.com/samber/lo"
	"github.com/tomill/plago/config"
	"github.com/tomill/plago/entry"
)

type Twitter struct {
	config.ExecParams
	client *twitter.Client
	userID string
}

func TwitterFetcher(c config.Config) (Fetcher, error) {
	client := oauth1.NewConfig(
		c.TwitterConsumerKey,
		c.TwitterConsumerSecret,
	).Client(
		context.WithValue(context.Background(), oauth1.HTTPClient, httpClient),
		oauth1.NewToken(c.TwitterToken, c.TwitterTokenSecret),
	)

	p := &Twitter{
		ExecParams: c.ExecParams,
		userID:     c.TwitterUserID,
		client: &twitter.Client{
			Authorizer: authorizer{},
			Client:     client,
			Host:       "https://api.twitter.com",
		},
	}

	return p, nil
}

var twOptions = struct {
	TweetFields []twitter.TweetField
	Expansions  []twitter.Expansion
	MediaFields []twitter.MediaField
}{
	TweetFields: []twitter.TweetField{
		twitter.TweetFieldCreatedAt,
		twitter.TweetFieldAuthorID,
		twitter.TweetFieldReferencedTweets,
		twitter.TweetFieldAttachments,
		twitter.TweetFieldEntities,
		twitter.TweetFieldInReplyToUserID,
		twitter.TweetFieldNoteTweet,
	},
	MediaFields: []twitter.MediaField{
		twitter.MediaFieldType,
		twitter.MediaFieldURL,
		twitter.MediaFieldPreviewImageURL,
	},
	Expansions: []twitter.Expansion{
		twitter.ExpansionAuthorID,
		twitter.ExpansionAttachmentsMediaKeys,
		twitter.ExpansionReferencedTweetsID,
		twitter.Expansion("referenced_tweets.id.attachments.media_keys"),
		twitter.ExpansionReferencedTweetsIDAuthorID,
	},
}

func (p Twitter) Fetch() (entry.Timeline, error) {
	timeline := entry.NewTimeline(p.ExecParams)

	res, err := p.client.UserTweetReverseChronologicalTimeline(context.Background(), p.userID,
		twitter.UserTweetReverseChronologicalTimelineOpts{
			StartTime:   p.Since,
			EndTime:     p.Until,
			TweetFields: twOptions.TweetFields,
			MediaFields: twOptions.MediaFields,
			Expansions:  twOptions.Expansions,
		},
	)
	if err != nil {
		return timeline, err
	}
	for _, tw := range res.Raw.TweetDictionaries() {
		timeline.Append(p.build(tw))
	}

	return timeline.Sorted(), nil
}

func (p Twitter) build(post *twitter.TweetDictionary) *entry.Entry {
	ts, _ := time.Parse(time.RFC3339, post.Tweet.CreatedAt)
	ts = ts.In(tz)
	if !timeinrange(ts, p.ExecParams) {
		return nil
	}

	e := &entry.Entry{
		Section:   ts.Format("2006-01-02 15:00"),
		Timestamp: ts,
		URL:       fmt.Sprintf("https://twitter.com/%s/status/%s", post.Author.UserName, post.Tweet.ID),
		User:      post.Author.Name,
		Text:      post.Tweet.Text,
		Reply:     post.Tweet.InReplyToUserID != "",
	}

	p.expand(e, post)

	for _, rt := range post.ReferencedTweets {
		switch rt.Reference.Type {
		case "retweeted":
			a := &entry.Entry{
				Text: fmt.Sprintf(`RT @%s: %s`, rt.TweetDictionary.Author.UserName, rt.TweetDictionary.Tweet.Text),
			}
			p.expand(a, rt.TweetDictionary)
			e.Text = a.Text
			e.Images = a.Images
			e.Attachments = a.Attachments
		case "quoted":
			a := e.AddAttachment(entry.Entry{
				Text: fmt.Sprintf(`%s: %s`, rt.TweetDictionary.Author.UserName, rt.TweetDictionary.Tweet.Text),
			})
			p.expand(a, rt.TweetDictionary)
		}
	}

	return e
}

func (p Twitter) expand(e *entry.Entry, post *twitter.TweetDictionary) {
	if note := post.Tweet.NoteTweet; note != nil {
		if strings.HasPrefix(e.Text, `RT @`) {
			e.Text = fmt.Sprintf(`RT @%s: `, post.Author.UserName)
		}
		e.Text = lo.Ellipsis(note.Text, 250)
	}

	if att := post.Tweet.Attachments; att != nil && len(att.PollIDs) > 0 {
		e.AddAttachment(entry.Entry{Text: "📊 [poll]"})
	}

	if ent := post.Tweet.Entities; ent != nil {
		for _, url := range ent.URLs {
			to := url.ExpandedURL
			if strings.HasPrefix(url.DisplayURL, "pic.x.com/") {
				to = ""
			} else if to == "https://twitter.com/i/web/status/"+post.Tweet.ID {
				to = ""
			} else if strings.HasPrefix(to, "https://twitter.com/") {
				for _, rt := range post.ReferencedTweets {
					if rt.Reference.Type == "quoted" && strings.HasSuffix(to, "/status/"+rt.Reference.ID) {
						to = ""
					}
				}
			} else if url.UnwoundURL != "" {
				to = url.UnwoundURL
			}

			e.Text = strings.ReplaceAll(e.Text, url.URL, to)

			if url.Title != "" {
				a := e.AddAttachment(entry.Entry{
					Text: url.Title,
				})
				if len(url.Images) > 1 {
					a.AddImage(url.Images[len(url.Images)-1].URL)
				}
			}
		}
	}

	for _, media := range post.AttachmentMedia {
		if media.PreviewImageURL != "" {
			e.AddImage(media.PreviewImageURL)
		} else if media.Type == "photo" && media.URL != "" {
			e.AddImage(media.URL)
		}
	}
}
