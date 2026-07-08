package output

import (
	"github.com/tomill/centre/entry"
)

type Flusher interface {
	Flush(entry.Timeline) error
}
