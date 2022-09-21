package main

import (
	"log"
	"time"

	"github.com/golobby/container/v3"
	"github.com/tomill/centre/config"
	"github.com/tomill/centre/input"
	"github.com/tomill/centre/output"
)

func main() {
	con := container.New()
	container.MustSingleton(con, config.GetOptions)

	for name, fn := range map[string]func(config.Config) input.Fetcher{
		"dummy":   input.DummyFetcher,
		"rtm":     input.RtmFetcher,
		"feed":    input.FeedFetcher,
		"twitter": input.TwitterFetcher,
	} {
		container.MustNamedTransientLazy(con, name, fn)
	}
	for name, fn := range map[string]func(config.Config) output.Flusher{
		"dump":  output.DumpFlusher,
		"gmail": output.GmailFlusher,
	} {
		container.MustNamedTransientLazy(con, name, fn)
	}

	err := con.Call(func(c config.Config) error {
		log.Printf("centre version %s", c.Version())
		var in input.Fetcher
		var out output.Flusher
		container.MustNamedResolve(con, &in, c.Input)
		container.MustNamedResolve(con, &out, c.Output)

		log.Printf("%s to %s; %d hour(s) %s - %s", c.Input, c.Output, c.Hours,
			c.Since.Format(time.RFC3339), c.Until.Format(time.RFC3339))
		timeline, err := in.Fetch()
		if err != nil {
			return err
		}

		log.Printf("fetched data from %s %d line(s). flush to %s...", c.Input, len(timeline.Messages), c.Output)
		if len(timeline.Messages) == 0 {
			return nil
		}

		return out.Flush(timeline)
	})
	if err != nil {
		log.Fatal(err)
	}
}
