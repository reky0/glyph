package mind

import (
	"bufio"
	"context"
	"encoding/json"
	"strings"
)

// ─── Claude (Anthropic) client ───────────────────────────────────────────────
//
// Anthropic's SSE format differs from OpenAI's:
//   - Each event is preceded by an "event: <type>" line.
//   - Text deltas arrive on "content_block_delta" events.
//   - The stream ends with a "message_stop" event (no "[DONE]" sentinel).

const anthropicAPIURL = "https://api.anthropic.com/v1/messages"
const anthropicVersion = "2023-06-01"
const claudeMaxTokens = 8192

type claudeClient struct {
	apiKey string
	model  string
}

type claudeRequest struct {
	Model     string         `json:"model"`
	MaxTokens int            `json:"max_tokens"`
	System    string         `json:"system"`
	Messages  []chatMessage  `json:"messages"`
	Stream    bool           `json:"stream"`
}

// SSE event payloads we care about.
type claudeContentBlockDelta struct {
	Delta struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"delta"`
}

func (c *claudeClient) Stream(ctx context.Context, system, user string) (<-chan string, error) {
	payload := claudeRequest{
		Model:     c.model,
		MaxTokens: claudeMaxTokens,
		System:    system,
		Messages:  []chatMessage{{Role: "user", Content: user}},
		Stream:    true,
	}

	body, err := doPost(ctx, anthropicAPIURL, map[string]string{
		"x-api-key":         c.apiKey,
		"anthropic-version": anthropicVersion,
	}, payload)
	if err != nil {
		return nil, err
	}

	ch := make(chan string, 64)
	go func() {
		defer close(ch)
		defer body.Close()

		scanner := bufio.NewScanner(body)
		var currentEvent string

		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			default:
			}

			line := scanner.Text()

			switch {
			case strings.HasPrefix(line, "event:"):
				currentEvent = strings.TrimSpace(strings.TrimPrefix(line, "event:"))

			case strings.HasPrefix(line, "data:"):
				payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))

				switch currentEvent {
				case "content_block_delta":
					var delta claudeContentBlockDelta
					if err := json.Unmarshal([]byte(payload), &delta); err != nil {
						continue
					}
					if delta.Delta.Type == "text_delta" && delta.Delta.Text != "" {
						select {
						case ch <- delta.Delta.Text:
						case <-ctx.Done():
							return
						}
					}

				case "message_stop":
					return
				}
			}
		}
	}()
	return ch, nil
}
