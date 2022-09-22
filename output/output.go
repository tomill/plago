package output

import (
	"encoding/json"
	"os"

	"github.com/tomill/centre/config"
	"github.com/tomill/centre/message"
)

type Flusher interface {
	Flush(message.Timeline) error
}

type Dump struct{}

func DumpFlusher(config.Config) Flusher {
	return &Dump{}
}

func (p Dump) Flush(timeline message.Timeline) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	return enc.Encode(timeline)
}
