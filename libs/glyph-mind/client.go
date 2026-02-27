package mind

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	core "github.com/reky0/glyph-core"
)

// Client can stream a response for a given system + user prompt.
type Client interface {
	// Stream sends a prompt and returns a channel of text chunks.
	// The caller must drain the channel. The channel is closed when
	// the stream ends. Any error after the channel is opened is
	// communicated by closing the channel; inspect the returned error
	// only for startup failures.
	Stream(ctx context.Context, system, user string) (<-chan string, error)
}

// NewClientFromConfig constructs the appropriate Client from configuration.
func NewClientFromConfig(cfg core.Config) (Client, error) {
	switch strings.ToLower(cfg.AIProvider) {
	case "ollama":
		return &ollamaClient{
			host:  cfg.OllamaHost,
			model: cfg.AIModel,
		}, nil
	case "groq", "":
		if cfg.APIKey == "" {
			return nil, &core.AppError{Msg: "api_key is required for groq provider"}
		}
		return &groqClient{
			apiKey: cfg.APIKey,
			model:  cfg.AIModel,
		}, nil
	case "claude":
		if cfg.APIKey == "" {
			return nil, &core.AppError{Msg: "api_key is required for claude provider"}
		}
		model := cfg.AIModel
		if model == "" {
			model = "claude-sonnet-4-6"
		}
		return &claudeClient{
			apiKey: cfg.APIKey,
			model:  model,
		}, nil
	default:
		return nil, &core.AppError{
			Msg: fmt.Sprintf("unknown ai_provider %q (valid: groq, ollama, claude)", cfg.AIProvider),
		}
	}
}

// ─── shared helpers ──────────────────────────────────────────────────────────

// chatMessage is the OpenAI-style message object used by both backends.
type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// sseStream reads an SSE response body, emitting text deltas to ch.
// It handles OpenAI-style `data: {...}` lines and closes ch when done.
func sseStream(ctx context.Context, body io.ReadCloser, ch chan<- string, extractDelta func([]byte) (string, bool, error)) {
	defer close(ch)
	defer body.Close()

	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
		}

		line := scanner.Text()
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if payload == "[DONE]" {
			return
		}
		delta, done, err := extractDelta([]byte(payload))
		if err != nil {
			// Silently skip malformed lines.
			continue
		}
		if done {
			return
		}
		if delta != "" {
			select {
			case ch <- delta:
			case <-ctx.Done():
				return
			}
		}
	}
}

// doPost sends a JSON POST request and returns the response body.
func doPost(ctx context.Context, url string, headers map[string]string, payload any) (io.ReadCloser, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("mind: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("mind: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("mind: request: %w", err)
	}
	if resp.StatusCode >= 400 {
		errBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("mind: server returned %d: %s", resp.StatusCode, string(errBody))
	}
	return resp.Body, nil
}

// ─── Groq client ─────────────────────────────────────────────────────────────

type groqClient struct {
	apiKey string
	model  string
}

type groqRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

type groqDelta struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
}

func (c *groqClient) Stream(ctx context.Context, system, user string) (<-chan string, error) {
	payload := groqRequest{
		Model: c.model,
		Messages: []chatMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
		Stream: true,
	}

	body, err := doPost(ctx, "https://api.groq.com/openai/v1/chat/completions",
		map[string]string{"Authorization": "Bearer " + c.apiKey},
		payload,
	)
	if err != nil {
		return nil, err
	}

	ch := make(chan string, 64)
	go sseStream(ctx, body, ch, func(data []byte) (string, bool, error) {
		var msg groqDelta
		if err := json.Unmarshal(data, &msg); err != nil {
			return "", false, err
		}
		if len(msg.Choices) == 0 {
			return "", false, nil
		}
		choice := msg.Choices[0]
		if choice.FinishReason != nil && *choice.FinishReason == "stop" {
			return "", true, nil
		}
		return choice.Delta.Content, false, nil
	})
	return ch, nil
}

// ─── Ollama client ────────────────────────────────────────────────────────────

type ollamaClient struct {
	host  string
	model string
}

type ollamaRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

type ollamaDelta struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
	Done bool `json:"done"`
}

func (c *ollamaClient) Stream(ctx context.Context, system, user string) (<-chan string, error) {
	host := c.host
	if host == "" {
		host = "http://localhost:11434"
	}
	url := strings.TrimRight(host, "/") + "/api/chat"

	payload := ollamaRequest{
		Model: c.model,
		Messages: []chatMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
		Stream: true,
	}

	body, err := doPost(ctx, url, nil, payload)
	if err != nil {
		return nil, err
	}

	// Ollama streams newline-delimited JSON, not SSE.
	ch := make(chan string, 64)
	go func() {
		defer close(ch)
		defer body.Close()

		scanner := bufio.NewScanner(body)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			default:
			}
			var msg ollamaDelta
			if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
				continue
			}
			if msg.Message.Content != "" {
				select {
				case ch <- msg.Message.Content:
				case <-ctx.Done():
					return
				}
			}
			if msg.Done {
				return
			}
		}
	}()
	return ch, nil
}
