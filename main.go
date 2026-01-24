package main

import (
	"fmt"
	"log"
	"runtime/debug"
	"time"

	"github.com/golobby/container/v3"
	"github.com/tomill/centre/config"
	"github.com/tomill/centre/input"
	"github.com/tomill/centre/output"
)

func main() {
	var msg string
	defer func() {
		if err := recover(); err != nil {
			log.Fatalf("[ERROR] panic: centre %s error: %s\n%s", msg, err, debug.Stack())
		}
	}()

	con := container.New()
	container.MustSingleton(con, config.GetOptions)

	for name, fn := range map[string]func(config.Config) (input.Fetcher, error){
		"dummy":    input.DummyFetcher,
		"raw":      input.RawFetcher,
		"stdin":    input.StdinFetcher,
		"rtm":      input.RtmFetcher,
		"feed":     input.FeedFetcher,
		"twitter":  input.TwitterFetcher,
		"xlist":    input.TwitterListFetcher,
		"bluesky":  input.BlueskyFetcher,
		"slack":    input.SlackFetcher,
		"slack_ch": input.SlackChannelsFetcher,
		"discord":  input.DiscordFetcher,
	} {
		container.MustNamedTransientLazy(con, name, fn)
	}
	for name, fn := range map[string]func(config.Config) output.Flusher{
		"dump":  output.DumpFlusher,
		"gmail": output.GmailFlusher,
	} {
		container.MustNamedTransientLazy(con, name, fn)
	}

	container.MustCall(con, func(c config.Config) error {
		msg = fmt.Sprintf("--in %s --out %s --since %q --until %q", c.Input, c.Output,
			c.Since.Format(time.RFC3339), c.Until.Format(time.RFC3339))

		log.Printf("initialising %s", msg)

		var in input.Fetcher
		var out output.Flusher
		container.MustNamedResolve(con, &in, c.Input)
		container.MustNamedResolve(con, &out, c.Output)

		timeline, err := in.Fetch()
		if err != nil {
			return err
		}

		if len(timeline.Messages) == 0 {
			log.Printf("fetched no data from %s. quit.", c.Input)
			return nil
		}

		log.Printf("fetched data from %s %d line(s). flush to %s ...", c.Input, len(timeline.Messages), c.Output)
		return out.Flush(timeline)
	})
}
