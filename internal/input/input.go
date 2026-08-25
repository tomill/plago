package input

import (
	"github.com/tomill/plago"
	"github.com/tomill/plago/internal/config"
)

type Fetcher interface {
	Fetch() (plago.Timeline, error)
}

var FetcherRegistry = map[string]func(config.Config) (Fetcher, error){
	"dummy":   DummyFetcher,
	"stdin":   StdinFetcher,
	"feed":    FeedFetcher,
	"twitter": TwitterFetcher,
	"twlist":  TwListFetcher,
	"bluesky": BlueskyFetcher,
	"slack":   SlackFetcher,
	"discord": DiscordFetcher,
	"youtube": YoutubeFetcher,
}
