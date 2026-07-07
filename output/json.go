package output

import (
	"encoding/json"
	"os"

	"github.com/tomill/centre/config"
	"github.com/tomill/centre/message"
)

type JSON struct {
	json *json.Encoder
}

func JSONFlusher(c config.Config) (Flusher, error) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)

	return &JSON{
		json: enc,
	}, nil
}

func (p JSON) Flush(timeline message.Timeline) error {
	return p.json.Encode(timeline)
}
