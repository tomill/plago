package input

import (
	"github.com/tomill/plago/entry"
)

type Fetcher interface {
	Fetch() (entry.Timeline, error)
}
