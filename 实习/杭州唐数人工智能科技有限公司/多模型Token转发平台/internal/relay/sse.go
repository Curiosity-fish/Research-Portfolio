package relay

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// StreamReader reads Server-Sent Events from src, forwards each complete data
// event to dst (followed by a flush if flusher is provided), and reports every
// data payload through onData.
//
// SSE format handled here follows the OpenAI streaming spec:
//   data: {...}\n\n
//   data: {...}\n\n
//   data: [DONE]\n\n
//
// Empty lines separate events. Lines starting with ":" are comments and ignored.
func StreamReader(dst io.Writer, src io.Reader, flusher http.Flusher, onData func(data string)) error {
	reader := bufio.NewReader(src)
	var eventLines []string

	flush := func() error {
		if flusher != nil {
			flusher.Flush()
		}
		return nil
	}

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				// Flush any remaining event before returning.
				if len(eventLines) > 0 {
					data := extractData(eventLines)
					if data != "" {
						if _, werr := fmt.Fprintf(dst, "data: %s\n\n", data); werr != nil {
							return werr
						}
						if ferr := flush(); ferr != nil {
							return ferr
						}
						if onData != nil {
							onData(data)
						}
					}
				}
				return nil
			}
			return err
		}

		// Trim trailing newline but keep the rest intact.
		line = bytes.TrimSuffix(line, []byte("\n"))
		if len(line) > 0 && line[len(line)-1] == '\r' {
			line = line[:len(line)-1]
		}

		if len(line) == 0 {
			// Empty line marks the end of an event.
			if len(eventLines) > 0 {
				data := extractData(eventLines)
				if data != "" {
					if _, werr := fmt.Fprintf(dst, "data: %s\n\n", data); werr != nil {
						return werr
					}
					if ferr := flush(); ferr != nil {
						return ferr
					}
					if onData != nil {
						onData(data)
					}
				}
				eventLines = eventLines[:0]
			}
			continue
		}

		// Ignore comments.
		if line[0] == ':' {
			continue
		}

		eventLines = append(eventLines, string(line))
	}
}

// extractData returns the concatenated data payload from SSE event lines.
func extractData(lines []string) string {
	var parts []string
	for _, line := range lines {
		if strings.HasPrefix(line, "data:") {
			parts = append(parts, strings.TrimSpace(strings.TrimPrefix(line, "data:")))
		}
	}
	return strings.Join(parts, "\n")
}

// Usage holds token counts parsed from a streaming usage chunk.
// It supports both OpenAI (`prompt_tokens`/`completion_tokens`) and Anthropic
// (`input_tokens`/`output_tokens`) field names.
type Usage struct {
	PromptTokens     int64 `json:"prompt_tokens"`
	CompletionTokens int64 `json:"completion_tokens"`
	InputTokens      int64 `json:"input_tokens"`
	OutputTokens     int64 `json:"output_tokens"`
}

// TotalTokens returns the sum of all parsed token counts.
func (u Usage) TotalTokens() int64 {
	input := u.PromptTokens + u.InputTokens
	output := u.CompletionTokens + u.OutputTokens
	return input + output
}

// ParseUsage scans collected SSE data payloads for the first usage object.
// OpenAI sends usage in the final data chunk before `[DONE]`.
func ParseUsage(data []string) Usage {
	for _, payload := range data {
		if payload == "[DONE]" {
			continue
		}
		payload = strings.TrimSpace(payload)

		var chunk struct {
			Usage *Usage `json:"usage"`
		}
		if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
			continue
		}
		if chunk.Usage != nil {
			return Usage{
				PromptTokens:     chunk.Usage.PromptTokens,
				CompletionTokens: chunk.Usage.CompletionTokens,
			}
		}
	}
	return Usage{}
}

// ParseAnthropicUsage scans collected SSE data payloads for the first Anthropic
// usage object. Anthropic uses `input_tokens`/`output_tokens` field names.
func ParseAnthropicUsage(data []string) Usage {
	for _, payload := range data {
		payload = strings.TrimSpace(payload)

		var chunk struct {
			Usage *Usage `json:"usage"`
		}
		if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
			continue
		}
		if chunk.Usage != nil {
			return Usage{
				InputTokens:  chunk.Usage.InputTokens,
				OutputTokens: chunk.Usage.OutputTokens,
			}
		}
	}
	return Usage{}
}
