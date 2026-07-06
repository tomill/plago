package input

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/tomill/centre/config"
	"github.com/tomill/centre/message"
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

func (p Stdin) Fetch() (message.Timeline, error) {
	var timeline message.Timeline

	dec := json.NewDecoder(os.Stdin)
	err := dec.Decode(&timeline)

	return timeline, err
}
