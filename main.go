package main

import (
	"fmt"
	"log"

	"github.com/tomill/centre/config"
	"github.com/tomill/centre/input"
	"github.com/tomill/centre/output"
)

func main() {
	c := config.GetOptions()

	if err := work(c); err != nil {
		log.Fatal(err)
	}
}

func work(c config.Config) error {
	log.Println(c)
	defer log.Println("done.")

	in := fetcher(c)
	if in == nil {
		return fmt.Errorf("failed to load input plugin: %q", c.Input)
	}

	out := flusher(c)
	if out == nil {
		return fmt.Errorf("failed to load output plugin: %q", c.Output)
	}

	timeline, err := in.Fetch()
	if err != nil {
		return fmt.Errorf("input plugin error: %w", err)
	}

	log.Printf("fetched from %s %d line(s). flush to %s...", c.Input, len(timeline.Messages), c.Output)

	if len(timeline.Messages) == 0 {
		return nil
	}
	if err := out.Flush(timeline); err != nil {
		return fmt.Errorf("output plugin error: %w", err)
	}

	return nil
}

func fetcher(c config.Config) input.Fetcher {
	switch c.Input {
	case "test":
		return input.Dummy{}
	case "rtm":
		return input.RtmFetcher(c)
	default:
		return nil
	}
}

func flusher(c config.Config) output.Flusher {
	switch c.Output {
	case "stdout":
		return &output.Stdout{}
	case "dump":
		return &output.Dump{}
	case "gmail":
		return output.GmailFlusher(c)
	default:
		return nil
	}
}
