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
	c := container.New()
	container.MustSingleton(c, config.GetOptions)

	for name, fn := range map[string]func(config.Config) input.Fetcher{
		"dummy":   input.DummyFetcher,
		"rtm":     input.RtmFetcher,
		"twitter": input.TwitterFetcher,
	} {
		container.MustNamedTransientLazy(c, name, fn)
	}
	for name, fn := range map[string]func(config.Config) output.Flusher{
		"stdout": output.StdoutFlusher,
		"dump":   output.DumpFlusher,
		"gmail":  output.GmailFlusher,
	} {
		container.MustNamedTransientLazy(c, name, fn)
	}

	err := c.Call(func(conf config.Config) error {
		var in input.Fetcher
		var out output.Flusher
		container.MustNamedResolve(c, &in, conf.Input)
		container.MustNamedResolve(c, &out, conf.Output)

		log.Printf("centre (%s) %s to %s", conf.Version(), conf.Input, conf.Output)
		defer log.Println("done.")

		log.Printf("about %d hour(s) %s - %s", conf.Hours, conf.Since().Format(time.RFC3339), conf.Until.Format(time.RFC3339))
		timeline, err := in.Fetch()
		if err != nil {
			return err
		}

		log.Printf("fetched data from %s %d line(s). flush to %s...", conf.Input, len(timeline.Messages), conf.Output)
		if len(timeline.Messages) == 0 {
			return nil
		}

		return out.Flush(timeline)
	})
	if err != nil {
		log.Fatal(err)
	}
}
