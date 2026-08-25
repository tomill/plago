package input

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/samber/lo"
	"github.com/tomill/plago"
	"github.com/tomill/plago/internal/config"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type Youtube struct {
	config.ExecParams
	client *youtube.Service
	target config.List
}

func YoutubeFetcher(c config.Config) (Fetcher, error) {
	p := &Youtube{
		ExecParams: c.ExecParams,
		client: lo.Must(youtube.NewService(context.Background(),
			option.WithAPIKey(c.YoutubeAPIKey),
		)),
		target: c.YoutubeChannelIDs,
	}

	return p, nil
}

func (p *Youtube) Fetch() (plago.Timeline, error) {
	timeline := newTimeline(p.ExecParams)

	for _, channelID := range p.target {
		ch, err := p.client.Channels.List([]string{"contentDetails"}).
			Id(channelID).Do()
		if err != nil {
			timeline.AppendError(err)
			continue
		}

		var videoIDs []string
		pl, err := p.client.PlaylistItems.List([]string{"snippet"}).
			PlaylistId(ch.Items[0].ContentDetails.RelatedPlaylists.Uploads).MaxResults(50).Do()
		if err != nil {
			timeline.AppendError(err)
			continue
		}
		for _, item := range pl.Items {
			ts, _ := time.Parse(time.RFC3339, item.Snippet.PublishedAt)
			if timeinrange(ts, p.ExecParams) {
				videoIDs = append(videoIDs, item.Snippet.ResourceId.VideoId)
			}
		}
		if len(videoIDs) == 0 {
			continue
		}

		vl, err := p.client.Videos.List([]string{"snippet", "contentDetails", "liveStreamingDetails"}).
			Id(videoIDs...).Do()
		if err != nil {
			timeline.AppendError(err)
			continue
		}
		for _, item := range vl.Items {
			timeline.Append(p.build(item))
		}
	}

	return timeline.Sorted(), nil
}

func (p *Youtube) build(video *youtube.Video) *plago.Entry {
	ts, _ := time.Parse(time.RFC3339, video.Snippet.PublishedAt)

	if video.Snippet.LiveBroadcastContent != "none" || video.LiveStreamingDetails != nil {
		return nil
	}
	if strings.Contains(strings.ToLower(video.Snippet.Title), "#shorts") {
		return nil
	}

	dur, _ := time.ParseDuration(
		strings.ToLower(strings.TrimPrefix(video.ContentDetails.Duration, "PT")),
	)
	if dur <= 90*time.Second {
		return nil
	}

	entry := &plago.Entry{
		Channel:   video.Snippet.ChannelTitle,
		Timestamp: ts.In(tz),
		URL:       fmt.Sprintf("https://www.youtube.com/watch?v=%s", video.Id),
		User:      dur.String(),
		Text:      video.Snippet.Title,
	}

	if find, ok := lo.Find([]any{
		video.Snippet.Thumbnails.Standard,
		video.Snippet.Thumbnails.High,
		video.Snippet.Thumbnails.Default,
	}, lo.IsNotNil); ok {
		entry.AddImage(find.(*youtube.Thumbnail).Url)
	}

	return entry
}
