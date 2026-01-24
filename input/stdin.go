package input

import (
	"encoding/json"
	"fmt"
	"github.com/tomill/centre/config"
	"github.com/tomill/centre/message"
	"os"
)

type Stdin struct {
}

func StdinFetcher(c config.Config) (Fetcher, error) {
	if fi, _ := os.Stdin.Stat(); fi.Mode()&os.ModeNamedPipe == 0 {
		return nil, fmt.Errorf("no pipe")
	}

	return &Stdin{}, nil
}

func (p Stdin) Fetch() (message.Timeline, error) {
	var timeline message.Timeline

	dec := json.NewDecoder(os.Stdin)
	err := dec.Decode(&timeline)

	return timeline, err
}
