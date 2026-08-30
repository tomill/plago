package output

import (
	"encoding/json"
	"os"

	"github.com/tomill/plago"
	"github.com/tomill/plago/internal/config"
)

type JSON struct {
	json *json.Encoder
}

func JSONFlusher(c config.Config) (Flusher, error) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)

	return &JSON{enc}, nil
}

func (p *JSON) Flush(timeline plago.Timeline) error {
	return p.json.Encode(timeline)
}
