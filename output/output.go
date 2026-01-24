package output

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/tomill/centre/config"
	"github.com/tomill/centre/message"
)

type Flusher interface {
	Flush(message.Timeline) error
}

type Dump struct {
	json *json.Encoder
}

func DumpFlusher(c config.Config) Flusher {
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)

	if c.RawData != "" {
		var d message.Timeline
		_ = json.NewDecoder(strings.NewReader(c.RawData)).Decode(&d)
		_ = enc.Encode(d)
	}

	return &Dump{
		json: enc,
	}
}

func (p Dump) Flush(timeline message.Timeline) error {
	return p.json.Encode(timeline)
}
