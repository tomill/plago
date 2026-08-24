package input

import (
	"log/slog"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/tomill/plago"
	"github.com/tomill/plago/internal/config"
)

var (
	httpClient *http.Client
	tz         = time.FixedZone("Asia/Tokyo", 9*60*60)
)

func init() {
	if loc, _ := time.LoadLocation(os.Getenv("TZ")); loc != nil {
		tz = loc
	}

	retry := retryablehttp.NewClient()
	retry.RetryMax = 2
	retry.RetryWaitMin = 2 * time.Second
	retry.Logger = hclog.New(&hclog.LoggerOptions{
		Name:       "http-client",
		Level:      hclog.Info,
		Output:     os.Stdout,
		JSONFormat: true,
	})

	httpClient = retry.StandardClient()

	if strings.ToLower(os.Getenv("LOG_LEVEL")) == "debug" {
		httpClient = &http.Client{
			Transport: &debugTransport{rt: httpClient.Transport},
		}
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

type debugTransport struct {
	rt http.RoundTripper
}

func (t *debugTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if dump, err := httputil.DumpRequest(req, true); err == nil {
		slog.Debug("==>", "http.Request", string(dump))
	}

	res, err := t.rt.RoundTrip(req)
	if err != nil {
		return res, err
	}

	if dump, err := httputil.DumpResponse(res, true); err == nil {
		slog.Debug("<--", "http.Response", string(dump))
	}
	return res, err
}
