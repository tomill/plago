package output

import (
	"github.com/tomill/centre/message"
)

type Flusher interface {
	Flush(message.Timeline) error
}
