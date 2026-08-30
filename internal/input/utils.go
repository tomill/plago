package input

import (
	"fmt"
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
		Entries: make([]plago.Entry, 0),
	}
}

func channelIDsEither(channelIDs config.List, channels []config.Channel) ([]string, error) {
	list := channelIDs
	if len(list) == 0 {
		for _, ch := range channels {
			list = append(list, ch.ChannelID)
		}
	}
	if len(list) == 0 {
		return nil, fmt.Errorf("A list of target channelIDs is required")
	}
	return list, nil
}

func timeinrange(ts time.Time, p config.ExecParams) bool {
	return !ts.Before(p.Since) && ts.Before(p.Until)
}

type cache[T any] map[string]T

func (c cache[T]) get(key string, callback func() T) T {
	if v, ok := (c)[key]; ok {
		return v
	}

	v := callback()
	(c)[key] = v
	return v
}
