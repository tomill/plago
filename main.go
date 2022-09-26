package main

import (
	"log"
	"runtime/debug"
	"time"

	"github.com/golobby/container/v3"
	"github.com/tomill/centre/config"
	"github.com/tomill/centre/input"
	"github.com/tomill/centre/output"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			log.Fatalf("[ERROR] panic: %s\n%s", err, debug.Stack())
		}
	}()

	con := container.New()
	container.MustSingleton(con, config.GetOptions)

	for name, fn := range map[string]func(config.Config) (input.Fetcher, error){
		"dummy":   input.DummyFetcher,
		"rtm":     input.RtmFetcher,
		"feed":    input.FeedFetcher,
		"twitter": input.TwitterFetcher,
		"slack":   input.SlackFetcher,
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
		log.Printf("initialising the %s to %s plugins. range: %s - %s", c.Input, c.Output,
			c.Since.Format(time.RFC3339), c.Until.Format(time.RFC3339))

		var in input.Fetcher
		var out output.Flusher
		container.MustNamedResolve(con, &in, c.Input)
		container.MustNamedResolve(con, &out, c.Output)

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
}
