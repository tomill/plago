package output

import (
	"encoding/json"
	"fmt"
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
	enc.SetIndent("", "  ")

	fmt.Println("\n// config")
	_ = enc.Encode(c)

	if c.RawData != "" {
		var d message.Timeline
		_ = json.NewDecoder(strings.NewReader(c.RawData)).Decode(&d)
		fmt.Println("\n// raw (decoded)")
		_ = enc.Encode(d)
	}

	return &Dump{
		json: enc,
	}
}

func (p Dump) Flush(timeline message.Timeline) error {
	fmt.Println("\n// timeline")
	return p.json.Encode(timeline)
}
