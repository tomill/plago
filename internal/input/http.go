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
)

var httpClient *http.Client

func init() {
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
