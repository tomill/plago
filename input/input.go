package input

import (
	"github.com/tomill/centre/message"
)

type Fetcher interface {
	Fetch() (message.Timeline, error)
}
