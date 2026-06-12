package scihub

import (
	"net/http"
	"time"

	enetxhttp "github.com/enetx/http"
	"github.com/enetx/surf"
)

// NewHTTPClient builds a Chrome-impersonating HTTP client matching the Surf
// setup used in internal/client/client.go.
func NewHTTPClient(timeout time.Duration) *http.Client {
	builder := surf.NewClient().
		Builder().
		Impersonate().
		Chrome().
		Timeout(timeout)
	builder = builder.ForceHTTP2().Session()
	surfClient := builder.Build().Unwrap()
	if t, ok := surfClient.GetTransport().(*enetxhttp.Transport); ok {
		t.ResponseHeaderTimeout = timeout
	}
	httpClient := surfClient.Std()
	httpClient.Timeout = timeout
	return httpClient
}
