package input

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/tomill/plago/config"
	"github.com/tomill/plago/entry"
)

type Stdin struct {
	config.ExecParams
}

func StdinFetcher(c config.Config) (Fetcher, error) {
	if fi, _ := os.Stdin.Stat(); fi.Mode()&os.ModeNamedPipe == 0 {
		return nil, fmt.Errorf("no pipe")
	}

	return &Stdin{c.ExecParams}, nil
}

func (p Stdin) Fetch() (entry.Timeline, error) {
	timeline := entry.Timeline{
		Source:  "stdin",
		Entries: make([]entry.Entry, 0),
	}

	err := json.NewDecoder(os.Stdin).Decode(&timeline)
	return timeline, err
}
