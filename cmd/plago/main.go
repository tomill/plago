package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"runtime/debug"

	"github.com/samber/lo"
	"github.com/tomill/plago/internal/config"
	"github.com/tomill/plago/internal/filter"
	"github.com/tomill/plago/internal/input"
	"github.com/tomill/plago/internal/output"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
		 	slog.Error(fmt.Sprintf("%v", err), "stack", string(debug.Stack()))
			os.Exit(1)
		}
	}()

	c := config.GetOptions()

	if err := run(c); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func run(c config.Config) error {
	fetcher, ok := input.FetcherRegistry[c.Input]
	if !ok {
		return fmt.Errorf("invalid --in %q (available %v)", c.Input, lo.Keys(input.FetcherRegistry))
	}

	flusher, ok := output.FlusherRegistry[c.Output]
	if !ok {
		return fmt.Errorf("invalid --out %q (available %v)", c.Output, lo.Keys(output.FlusherRegistry))
	}

	filter, ok := filter.New(c.Filter)
	if !ok {
		return fmt.Errorf("invalid --filter %q", c.Filter)
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

	log.Printf("plago fetched %d entries from %s", len(timeline.Entries), c.Input)
	if len(timeline.Entries) == 0 {
		return nil
	}

	if filter != nil {
		log.Printf("plago filters entries by %s", c.Filter)
		filter.Filter(&timeline)
	}

	log.Printf("plago flushing %d entries to %s ...", len(timeline.Entries), c.Output)
	return out.Flush(timeline)
}
