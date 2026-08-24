package output

import (
	"github.com/tomill/plago"
	"github.com/tomill/plago/internal/config"
)

type Flusher interface {
	Flush(plago.Timeline) error
}

var FlusherRegistry = map[string]func(config.Config) (Flusher, error){
	"json":  JSONFlusher,
	"gmail": GmailFlusher,
}
