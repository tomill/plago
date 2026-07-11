package input

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/tomill/centre/config"
	"github.com/tomill/centre/entry"
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
	var timeline entry.Timeline
	err := json.NewDecoder(os.Stdin).Decode(&timeline)
	return timeline, err
}
