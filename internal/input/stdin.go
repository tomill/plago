package input

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/tomill/plago"
	"github.com/tomill/plago/internal/config"
)

type Stdin struct {
	config.ExecParams
}

func StdinFetcher(c config.Config) (Fetcher, error) {
	if fi, _ := os.Stdin.Stat(); fi.Mode()&os.ModeCharDevice != 0 {
		return nil, fmt.Errorf("no pipe")
	}

	return &Stdin{c.ExecParams}, nil
}

func (p *Stdin) Fetch() (plago.Timeline, error) {
	timeline := newTimeline(p.ExecParams)
	err := json.NewDecoder(os.Stdin).Decode(&timeline)
	return timeline, err
}
