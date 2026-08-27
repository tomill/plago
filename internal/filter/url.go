package filter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/samber/lo"
	"github.com/tomill/plago"
	"golang.org/x/sync/errgroup"
)

type APIFilter struct {
	URL *url.URL
}

func (p *APIFilter) Filter(timeline *plago.Timeline) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	eg, ctx := errgroup.WithContext(ctx)
	eg.SetLimit(5)

	for i, entry := range timeline.Entries {
		eg.Go(func() error {
			slog.Debug("filter by API ==>", "entry.URL", entry.URL)
			updated, err := p.process(ctx, entry)
			if err != nil {
				slog.Debug("filter by API <-- ng", "entry.URL", entry.URL, "err", err)
				return nil
			}

			slog.Debug("filter by API <-- ok", "entry.URL", entry.URL)
			if updated == nil {
				timeline.Entries[i] = plago.Entry{}
			} else {
				timeline.Entries[i] = *updated
			}
			return nil
		})
	}

	_ = eg.Wait()

	timeline.Entries = lo.Filter(timeline.Entries, func(e plago.Entry, _ int) bool {
		return !e.IsZero()
	})
}

func (p *APIFilter) process(ctx context.Context, entry plago.Entry) (*plago.Entry, error) {
	var payload bytes.Buffer
	_ = json.NewEncoder(&payload).Encode(entry)

	req, _ := http.NewRequestWithContext(ctx, "POST", p.URL.String(), &payload)
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	} else if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", res.StatusCode)
	}
	defer res.Body.Close()

	var updated plago.Entry
	if err := json.NewDecoder(res.Body).Decode(&updated); err != nil {
		return nil, err
	}

	return &updated, nil
}
