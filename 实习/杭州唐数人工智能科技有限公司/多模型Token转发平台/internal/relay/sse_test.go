package relay

import (
	"bytes"
	"net/http/httptest"
	"testing"
)

func TestStreamReader_ForwardsEvents(t *testing.T) {
	src := bytes.NewReader([]byte("data: hello\n\ndata: world\n\n"))
	dst := httptest.NewRecorder()

	var captured []string
	err := StreamReader(dst, src, dst, func(data string) {
		captured = append(captured, data)
	})
	if err != nil {
		t.Fatalf("stream reader: %v", err)
	}

	if len(captured) != 2 {
		t.Fatalf("expected 2 events, got %d", len(captured))
	}
	if captured[0] != "hello" || captured[1] != "world" {
		t.Errorf("unexpected captured data: %v", captured)
	}
	if dst.Body.String() != "data: hello\n\ndata: world\n\n" {
		t.Errorf("unexpected output: %q", dst.Body.String())
	}
}

func TestStreamReader_IgnoresComments(t *testing.T) {
	src := bytes.NewReader([]byte(": comment\ndata: event\n\n"))
	dst := httptest.NewRecorder()

	var captured []string
	err := StreamReader(dst, src, dst, func(data string) {
		captured = append(captured, data)
	})
	if err != nil {
		t.Fatalf("stream reader: %v", err)
	}

	if len(captured) != 1 || captured[0] != "event" {
		t.Errorf("unexpected captured data: %v", captured)
	}
}

func TestParseUsage(t *testing.T) {
	data := []string{
		`{"id":"chatcmpl-1","choices":[{"delta":{"role":"assistant"}}]}`,
		`{"id":"chatcmpl-1","choices":[],"usage":{"prompt_tokens":5,"completion_tokens":5,"total_tokens":10}}`,
		"[DONE]",
	}

	usage := ParseUsage(data)
	if usage.PromptTokens != 5 || usage.CompletionTokens != 5 || usage.TotalTokens() != 10 {
		t.Errorf("unexpected usage: %+v", usage)
	}
}

func TestParseUsageNoUsage(t *testing.T) {
	data := []string{
		`{"id":"chatcmpl-1"}`,
		"[DONE]",
	}
	usage := ParseUsage(data)
	if usage.TotalTokens() != 0 {
		t.Errorf("expected zero usage, got %+v", usage)
	}
}
