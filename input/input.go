package input

import (
	"github.com/tomill/centre/entry"
)

type Fetcher interface {
	Fetch() (entry.Timeline, error)
}
