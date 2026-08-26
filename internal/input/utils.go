package input

import (
	"os"
	"time"

	"github.com/tomill/plago"
	"github.com/tomill/plago/internal/config"
)

var tz = time.FixedZone("Asia/Tokyo", 9*60*60)

func init() {
	if loc, _ := time.LoadLocation(os.Getenv("TZ")); loc != nil {
		tz = loc
	}
}

func newTimeline(c config.ExecParams) plago.Timeline {
	return plago.Timeline{
		Source:  c.Input,
		Subject: c.Subject,
		RefID:   c.RefID,
		Entries: []plago.Entry{},
	}
}

func timeinrange(ts time.Time, p config.ExecParams) bool {
	return !ts.Before(p.Since) && ts.Before(p.Until)
}

type cache[T any] map[string]T

func (c *cache[T]) get(key string, callback func() T) T {
	if v, ok := (*c)[key]; ok {
		return v
	}

	v := callback()
	(*c)[key] = v
	return v
}
