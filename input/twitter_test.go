package input

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/dghubble/oauth1"
	"github.com/g8rswimmer/go-twitter/v2"
	"github.com/tomill/plago/config"
)

func TestTwitter(t *testing.T) {
	p := &Twitter{
		ExecParams: config.ExecParams{
			Since: time.Unix(0, 0),
			Until: time.Now(),
		},
		client: &twitter.Client{
			Authorizer: authorizer{},
			Client: oauth1.NewConfig(
				os.Getenv("TWITTER_CONSUMER_KEY"),
				os.Getenv("TWITTER_CONSUMER_SECRET"),
			).Client(
				context.Background(),
				oauth1.NewToken(os.Getenv("TWITTER_OAUTH1_TOKEN"), os.Getenv("TWITTER_OAUTH1_TOKEN_SECRET")),
			),
			Host: "https://api.twitter.com",
		},
	}

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
