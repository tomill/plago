package input

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/tomill/plago"
	"github.com/tomill/plago/internal/config"
)

type URL struct {
	config.ExecParams
	url *url.URL
}

func URLFetcher(c config.Config) (Fetcher, error) {
	if c.URL == nil || (c.URL.Scheme != "http" && c.URL.Scheme != "https") {
		return nil, fmt.Errorf("invalid --url %q", c.URL)
	}

	p := &URL{
		ExecParams: c.ExecParams,
		url:        c.URL,
	}

	return p, nil
}

func (p *URL) Fetch() (plago.Timeline, error) {
	timeline := newTimeline(p.ExecParams)

	req, _ := http.NewRequest("GET", p.url.String(), nil)
	res, err := httpClient.Do(req)
	if err != nil {
		return timeline, fmt.Errorf("fetch failed: %w", err)
	}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		return timeline, fmt.Errorf("Status: %d", res.StatusCode)
	}

	switch ct := res.Header.Get("Content-Type"); {
	default:
		return timeline, fmt.Errorf("Unsupported Content-Type: %s", ct)
	case strings.HasPrefix(ct, "application/json"):
		if err = p.json(&timeline, res); err != nil {
			return timeline, err
		}
	}

	return timeline, nil
}

func (p *URL) json(timeline *plago.Timeline, res *http.Response) error {
	return json.NewDecoder(res.Body).Decode(timeline)
}
