package input

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/samber/lo"
	"github.com/tomill/plago"
	"github.com/tomill/plago/internal/config"
	"golang.org/x/net/publicsuffix"
)

type Bluesky struct {
	config.ExecParams
	client *xrpc.Client
}

func BlueskyFetcher(c config.Config) (Fetcher, error) {
	p := &Bluesky{
		ExecParams: c.ExecParams,
		client: &xrpc.Client{
			Host:   "https://bsky.social",
			Client: httpClient,
		},
	}

	out, err := atproto.ServerCreateSession(context.Background(), p.client, &atproto.ServerCreateSession_Input{
		Identifier: c.BlueskyHandle,
		Password:   c.BlueskyAppPassword,
	})
	if err != nil {
		return p, err
	}
	p.client.Auth = &xrpc.AuthInfo{
		AccessJwt:  out.AccessJwt,
		RefreshJwt: out.RefreshJwt,
		Handle:     out.Handle,
		Did:        out.Did,
	}

	return p, nil
}

func (p Bluesky) Fetch() (plago.Timeline, error) {
	timeline := newTimeline(p.ExecParams)

	res, err := bsky.FeedGetTimeline(context.Background(), p.client, "reverse-chronological", "", 100)
	if err != nil {
		return timeline, err
	}
	for _, feed := range res.Feed {
		timeline.Append(p.build(feed))
	}

	return timeline.Sorted(), nil
}

func (p Bluesky) build(feed *bsky.FeedDefs_FeedViewPost) *plago.Entry {
	post, ok := feed.Post.Record.Val.(*bsky.FeedPost)
	if !ok {
		return nil
	}

	ts, _ := time.Parse(time.RFC3339, post.CreatedAt)
	ts = ts.In(tz)
	if !timeinrange(ts, p.ExecParams) {
		return nil
	}

	entry := &plago.Entry{
		Section:   ts.Format("2006-01-02 15:00"),
		Timestamp: ts,
		URL: fmt.Sprintf("https://bsky.app/profile/%s/post%s",
			feed.Post.Author.Handle,
			feed.Post.Uri[strings.LastIndex(feed.Post.Uri, "/"):],
		),
		User: *feed.Post.Author.DisplayName,
		Text: p.text(post),
	}

	if feed.Reply != nil && feed.Reply.Parent.FeedDefs_PostView.Author.Handle != feed.Post.Author.Handle {
		entry.Reply = true
		entry.Text = fmt.Sprintf(`@%s %s`, p.handle(feed.Reply.Parent.FeedDefs_PostView.Author.Handle), entry.Text)
	}

	if feed.Reason != nil && feed.Reason.FeedDefs_ReasonRepost != nil {
		entry.Text = fmt.Sprintf(`RT @%s: %s`, p.handle(feed.Post.Author.Handle), entry.Text)
		entry.User = *feed.Reason.FeedDefs_ReasonRepost.By.DisplayName
	}

	if embed := feed.Post.Embed; embed != nil {
		if find, ok := lo.Find([]any{
			embed.EmbedImages_View,
			embed.EmbedVideo_View,
			embed.EmbedExternal_View,
			embed.EmbedRecord_View,
			embed.EmbedRecordWithMedia_View,
		}, lo.IsNotNil); ok {
			p.embed(entry, find)
		}
	}

	return entry
}

func (p Bluesky) embed(entry *plago.Entry, embed any) {
	switch v := embed.(type) {
	case *bsky.EmbedImages_View:
		for _, media := range v.Images {
			entry.AddImage(media.Thumb)
		}
	case *bsky.EmbedVideo_View:
		if v.Thumbnail != nil {
			entry.AddImage(*v.Thumbnail)
		}
	case *bsky.EmbedExternal_View:
		a := entry.AddAttachment(plago.Entry{Text: v.External.Title})
		if !strings.Contains(entry.Text, v.External.Uri) {
			a.URL = v.External.Uri
		}
		if v.External.Thumb != nil {
			a.AddImage(*v.External.Thumb)
		}

	case *bsky.EmbedRecord_View:
		if v.Record != nil && v.Record.EmbedRecord_ViewRecord != nil {
			p.quoted(entry, v.Record.EmbedRecord_ViewRecord)
		}
	case *bsky.EmbedRecordWithMedia_View:
		if find, ok := lo.Find([]any{
			v.Media.EmbedImages_View,
			v.Media.EmbedVideo_View,
			v.Media.EmbedExternal_View,
		}, lo.IsNotNil); ok {
			p.embed(entry, find)
		}
		if v.Record != nil && v.Record.Record != nil && v.Record.Record.EmbedRecord_ViewRecord != nil {
			p.quoted(entry, v.Record.Record.EmbedRecord_ViewRecord)
		}
	}
}

func (p Bluesky) quoted(entry *plago.Entry, v *bsky.EmbedRecord_ViewRecord) {
	post, ok := v.Value.Val.(*bsky.FeedPost)
	if !ok {
		return
	}

	a := entry.AddAttachment(plago.Entry{
		Text: fmt.Sprintf(`%s: %s`, p.handle(v.Author.Handle), post.Text),
	})
	for _, v := range v.Embeds {
		if find, ok := lo.Find([]any{
			v.EmbedImages_View,
			v.EmbedVideo_View,
			v.EmbedExternal_View,
		}, lo.IsNotNil); ok {
			p.embed(a, find)
		}
	}
}

func (p Bluesky) handle(s string) string {
	if name, ok := strings.CutSuffix(s, ".bsky.social"); ok {
		return name
	}

	if tld, _ := publicsuffix.PublicSuffix(s); tld != "" {
		return strings.TrimSuffix(s, "."+tld)
	}

	return s
}

func (p Bluesky) text(post *bsky.FeedPost) string {
	text := post.Text
	if len(post.Facets) == 0 {
		return text
	}

	var buff strings.Builder
	var lastEnd int64
	for _, f := range post.Facets {
		start, end := f.Index.ByteStart, f.Index.ByteEnd
		link := ""
		for _, feat := range f.Features {
			if feat.RichtextFacet_Link != nil {
				link = feat.RichtextFacet_Link.Uri
				break
			}
		}

		buff.WriteString(text[lastEnd:start])
		buff.WriteString(link)
		lastEnd = end
	}
	buff.WriteString(text[lastEnd:])

	return buff.String()
}
