package filter

import (
	"net/url"

	"github.com/tomill/plago"
)

type Filter interface {
	Filter(*plago.Timeline)
}

func New(filter string) (Filter, bool) {
	if filter == "" {
		return nil, true
	}

	if u, err := url.Parse(filter); err == nil && u.Scheme == "https" {
		return &APIFilter{URL: *u}, true
	}

	return nil, false
}
