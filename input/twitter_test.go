package input

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/g8rswimmer/go-twitter/v2"
	"github.com/tomill/plago/config"
)

func TestTwitter(t *testing.T) {
	f, err := TwitterFetcher(config.Config{
		TwitterConsumerKey:    os.Getenv("TWITTER_CONSUMER_KEY"),
		TwitterConsumerSecret: os.Getenv("TWITTER_CONSUMER_SECRET"),
		TwitterToken:          os.Getenv("TWITTER_OAUTH1_TOKEN"),
		TwitterTokenSecret:    os.Getenv("TWITTER_OAUTH1_TOKEN_SECRET"),
	})
	if err != nil {
		t.Fatal(err)
	}
	p := f.(*Twitter)

	res, err := p.client.TweetLookup(context.Background(), []string{
		"2076489884290949245",
	},
		twitter.TweetLookupOpts{
			TweetFields: twOptions.TweetFields,
			MediaFields: twOptions.MediaFields,
			Expansions:  twOptions.Expansions,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	_ = json.NewEncoder(os.Stderr).Encode(res)

	for _, tw := range res.Raw.TweetDictionaries() {
		_ = json.NewEncoder(os.Stderr).Encode(tw)
		e := p.build(tw)
		_ = json.NewEncoder(os.Stderr).Encode(e)
	}
}
