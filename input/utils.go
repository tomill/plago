package input

import (
	"log/slog"
	"net/http"
	"net/http/httputil"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/tomill/plago/config"
)

var (
	tz               *time.Location
	httpClient       *http.Client
	reAmazon         = regexp.MustCompile(`^(https://www.amazon.co.jp)/?[^/]*/[^/]*/([A-Z0-9]{10}).*$`)
	reUtmTracker     = regexp.MustCompile(`[#?&]utm_[a-z0-9]+=[^&]+`)
	reMarkdownEscape = regexp.MustCompile(`\\([_*\[\]()~` + "`" + `>#+\-=|{}.!])`)
)

func init() {
	if loc, _ := time.LoadLocation(os.Getenv("TZ")); loc != nil {
		tz = loc
	} else {
		tz = time.FixedZone("Asia/Tokyo", 9*60*60)
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

func timeinrange(ts time.Time, p config.ExecParams) bool {
	return !ts.Before(p.Since) && ts.Before(p.Until)
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

type authorizer struct{}

func (a authorizer) Add(*http.Request) {} // no-op. added by oauth1.Client

type cache[T any] map[string]T

func (c *cache[T]) get(key string, callback func() T) T {
	if v, ok := (*c)[key]; ok {
		return v
	}

	v := callback()
	(*c)[key] = v
	return v
}
