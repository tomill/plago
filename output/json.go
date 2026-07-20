package output

import (
	"encoding/json"
	"os"

	"github.com/tomill/plago/config"
	"github.com/tomill/plago/entry"
)

type JSON struct {
	json *json.Encoder
}

func JSONFlusher(c config.Config) (Flusher, error) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)

	return &JSON{enc}, nil
}

func (p JSON) Flush(timeline entry.Timeline) error {
	return p.json.Encode(timeline)
}
