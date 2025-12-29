// Package httpclient provides HTTP client utilities with configurable timeouts
// and redirect behavior for YouTube Shorts detection.
package httpclient

import (
	"net/http"
	"time"
)

// HTTPClient defines the interface for HTTP operations.
// This interface allows for easy mocking in tests.
type HTTPClient interface {
	Get(url string) (*http.Response, error)
	Head(url string) (*http.Response, error)
	Do(req *http.Request) (*http.Response, error)
}

// NewHTTPClient creates a standard HTTP client with the specified timeout.
// This client follows redirects by default.
func NewHTTPClient(timeout time.Duration) HTTPClient {
	return &http.Client{
		Timeout: timeout,
	}
}

// NewNoRedirectClient creates an HTTP client that does not follow redirects.
// This is useful for detecting YouTube Shorts URLs which redirect from
// /watch?v=ID to /shorts/ID.
func NewNoRedirectClient(timeout time.Duration) HTTPClient {
	return &http.Client{
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}
