package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"runtime/debug"

	"github.com/samber/lo"
	"github.com/tomill/plago/internal/config"
	"github.com/tomill/plago/internal/input"
	"github.com/tomill/plago/internal/output"
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

func run(c config.Config) error {
	fetcher, ok := input.FetcherRegistry[c.Input]
	if !ok {
		return fmt.Errorf("invalid --in %q (available %v)", c.Input, lo.Keys(input.FetcherRegistry))
	}

	flusher, ok := output.FlusherRegistry[c.Output]
	if !ok {
		return fmt.Errorf("invalid --out %q (available %v)", c.Output, lo.Keys(output.FlusherRegistry))
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

	if len(timeline.Entries) == 0 {
		log.Printf("plago fetched 0 entry from %s", c.Input)
		return nil
	}

	log.Printf("plago fetched %d entries from %s. flushing them to %s ...", len(timeline.Entries), c.Input, c.Output)
	return out.Flush(timeline)
}
