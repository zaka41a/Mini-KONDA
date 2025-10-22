package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Ollama represents the client for communicating with the Ollama API.
type Ollama struct {
	baseURL string
	model   string
	client  *http.Client
}

// NewOllama initializes a new Ollama client.
func NewOllama(baseURL, model string) *Ollama {
	return &Ollama{
		baseURL: strings.TrimRight(baseURL, "/"),
		model:   model,
		client:  &http.Client{Timeout: 600 * time.Second}, // 10 minutes
	}
}

// DescribeColumn calls Ollama (non-streaming) and retries if necessary.
func (o *Ollama) DescribeColumn(ctx context.Context, column string) string {
	body := map[string]interface{}{
		"model":  o.model,
		"prompt": fmt.Sprintf("Describe the semantic meaning and role of the dataset column: %q", column),
		"stream": false,
	}
	payload, _ := json.Marshal(body)

	for attempt := 0; attempt < 3; attempt++ {
		req, _ := http.NewRequestWithContext(ctx, "POST", o.baseURL+"/api/generate", bytes.NewBuffer(payload))
		req.Header.Set("Content-Type", "application/json")

		resp, err := o.client.Do(req)
		if err != nil {
			fmt.Printf("⚠️ Ollama request failed (attempt %d): %v\n", attempt+1, err)
			time.Sleep(time.Duration(1+attempt) * time.Second)
			continue
		}

		b, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()

		var out map[string]interface{}
		if err := json.Unmarshal(b, &out); err == nil {
			if s, ok := out["response"].(string); ok && strings.TrimSpace(s) != "" {
				return s
			}
		}

		fmt.Printf("⚠️ Ollama returned empty or invalid response (attempt %d)\n", attempt+1)
		time.Sleep(time.Duration(1+attempt) * time.Second)
	}

	return "⚠️ Ollama not reachable"
}

// StreamDescribeColumn streams responses from Ollama token by token.
func (o *Ollama) StreamDescribeColumn(ctx context.Context, column string, onToken func(token string)) {
	body := map[string]interface{}{
		"model":  o.model,
		"prompt": fmt.Sprintf("Describe the semantic meaning and role of the dataset column: %q", column),
		"stream": true,
	}
	payload, _ := json.Marshal(body)

	req, _ := http.NewRequestWithContext(ctx, "POST", o.baseURL+"/api/generate", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		onToken("⚠️ Ollama not reachable")
		return
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	for {
		var chunk struct {
			Response string `json:"response"`
			Done     bool   `json:"done"`
		}
		if err := dec.Decode(&chunk); err != nil {
			break
		}
		if chunk.Response != "" {
			onToken(chunk.Response)
		}
		if chunk.Done {
			break
		}
	}
}
