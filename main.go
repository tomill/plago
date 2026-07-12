package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"runtime/debug"

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

	conf := config.GetOptions()

	if err := run(conf); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func run(c config.Config) error {
	fetcher, ok := map[string]func(config.Config) (input.Fetcher, error){
		"dummy":    input.DummyFetcher,
		"stdin":    input.StdinFetcher,
		"feed":     input.FeedFetcher,
		"twitter":  input.TwitterFetcher,
		"twlist":   input.TwListFetcher,
		"togetter": input.TogetterFetcher,
		"bluesky":  input.BlueskyFetcher,
		"slack":    input.SlackFetcher,
		"discord":  input.DiscordFetcher,
	}[c.Input]
	if !ok {
		return fmt.Errorf("invalid --in: %s", c.Input)
	}

	flusher, ok := map[string]func(config.Config) (output.Flusher, error){
		"json":  output.JSONFlusher,
		"gmail": output.GmailFlusher,
	}[c.Output]
	if !ok {
		return fmt.Errorf("invalid --out: %s", c.Output)
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
