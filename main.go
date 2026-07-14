package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"runtime/debug"

	"github.com/samber/lo"
	"github.com/tomill/centre/config"
	"github.com/tomill/centre/input"
	"github.com/tomill/centre/output"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			slog.Error(err.(error).Error(), "stack", string(debug.Stack()))
			os.Exit(1)
		}
	}()

	c := config.GetOptions()

	if err := run(c); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

var fetcherRegistry = map[string]func(config.Config) (input.Fetcher, error){
	"dummy":   input.DummyFetcher,
	"stdin":   input.StdinFetcher,
	"feed":    input.FeedFetcher,
	"twitter": input.TwitterFetcher,
	"twlist":  input.TwListFetcher,
	"bluesky": input.BlueskyFetcher,
	"slack":   input.SlackFetcher,
	"discord": input.DiscordFetcher,
}

var flusherRegistry = map[string]func(config.Config) (output.Flusher, error){
	"json":  output.JSONFlusher,
	"gmail": output.GmailFlusher,
}

func run(c config.Config) error {
	fetcher, ok := fetcherRegistry[c.Input]
	if !ok {
		return fmt.Errorf("invalid --in %q (available %v)", c.Input, lo.Keys(fetcherRegistry))
	}

	flusher, ok := flusherRegistry[c.Output]
	if !ok {
		return fmt.Errorf("invalid --out %q (available %v)", c.Output, lo.Keys(flusherRegistry))
	}

	in, err := fetcher(c)
	if err != nil {
		return err
	}

	out, err := flusher(c)
	if err != nil {
		return err
	}

	log.Printf("plago %s", c.ExecParams)
	timeline, err := in.Fetch()
	if err != nil {
		return err
	}

	log.Printf("plago fetched %d entries from %s. flushes them to %s ...", len(timeline.Entries), c.Input, c.Output)
	return out.Flush(timeline)
}
