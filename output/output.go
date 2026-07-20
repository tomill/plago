package output

import (
	"github.com/tomill/plago/entry"
)

type Flusher interface {
	Flush(entry.Timeline) error
}
