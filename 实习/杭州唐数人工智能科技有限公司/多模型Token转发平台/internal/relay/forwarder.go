package relay

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// Forwarder executes an upstream HTTP request and returns the raw response.
// The caller is responsible for closing the response body.
type Forwarder interface {
	Forward(ctx context.Context, req *http.Request) (*http.Response, error)
}

type forwarder struct {
	client *http.Client
}

// NewForwarder creates a Forwarder with the given request timeout.
func NewForwarder(timeout time.Duration) Forwarder {
	return &forwarder{
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// Forward sends the request using a timeout-bound HTTP client.
func (f *forwarder) Forward(ctx context.Context, req *http.Request) (*http.Response, error) {
	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("forward request: %w", err)
	}
	return resp, nil
}
