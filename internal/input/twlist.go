package input

import (
	"context"
	"time"

	"github.com/g8rswimmer/go-twitter/v2"
	"github.com/samber/lo"
	"github.com/tomill/plago"
	"github.com/tomill/plago/internal/config"
)

type TwList struct {
	config.ExecParams
	proxy  *Twitter
	target config.List
}

func TwListFetcher(c config.Config) (Fetcher, error) {
	p := &TwList{
		ExecParams: c.ExecParams,
		proxy:      lo.Must(TwitterFetcher(c)).(*Twitter),
		target:     c.TwitterListIDs,
	}

	return p, nil
}

func (p *TwList) Fetch() (plago.Timeline, error) {
	timeline := newTimeline(p.ExecParams)

	for _, listID := range p.target {
		var channel string
		if res, err := p.proxy.client.ListLookup(context.Background(), listID, twitter.ListLookupOpts{}); err != nil {
			timeline.AppendError(err)
			continue
		} else {
			channel = res.Raw.List.Name
		}

		var page string
		for {
			res, err := p.proxy.client.ListTweetLookup(context.Background(), listID,
				twitter.ListTweetLookupOpts{
					MaxResults:      10,
					PaginationToken: page,
					TweetFields:     twOptions.TweetFields,
					MediaFields:     twOptions.MediaFields,
					Expansions:      twOptions.Expansions,
				},
			)
			if err != nil {
				return timeline, err
			}

			var quit bool
			for _, tw := range res.Raw.TweetDictionaries() {
				ts, _ := time.Parse(time.RFC3339, tw.Tweet.CreatedAt)
				if ts.Before(p.Since) {
					quit = true
					break
				}

				if entry := p.proxy.build(tw); entry != nil {
					entry.Channel = channel
					timeline.Append(entry)
				}
			}

			if quit {
				break
			} else {
				page = res.Meta.NextToken
			}
		}
	}

	return timeline.Sorted(), nil
}
