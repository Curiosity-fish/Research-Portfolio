package relay

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestForwarder_Forward(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer sk-upstream" {
			t.Errorf("unexpected authorization header: %s", r.Header.Get("Authorization"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("unexpected content-type: %s", r.Header.Get("Content-Type"))
		}

		body, _ := io.ReadAll(r.Body)
		if string(body) != `{"model":"gpt-4o"}` {
			t.Errorf("unexpected body: %s", string(body))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"chatcmpl-1"}`))
	}))
	defer upstream.Close()

	fwd := NewForwarder(5 * time.Second)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, upstream.URL, bytes.NewReader([]byte(`{"model":"gpt-4o"}`)))
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer sk-upstream")
	req.Header.Set("Content-Type", "application/json")

	resp, err := fwd.Forward(context.Background(), req)
	if err != nil {
		t.Fatalf("forward: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != `{"id":"chatcmpl-1"}` {
		t.Errorf("unexpected response body: %s", string(body))
	}
}

func TestForwarder_ForwardTimeout(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	fwd := NewForwarder(10 * time.Millisecond)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, upstream.URL, http.NoBody)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	_, err = fwd.Forward(context.Background(), req)
	if err == nil {
		t.Fatal("expected timeout error")
	}
}
