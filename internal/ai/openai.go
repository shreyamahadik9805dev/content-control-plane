package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/shreyafeo/content-control-plane/internal/domain"
)

type openaiClient struct {
	apiKey  string
	model   string
	baseURL string
	client  *http.Client
}

// NewOpenAI calls OpenAI Chat Completions with JSON response format.
func NewOpenAI(apiKey, model string, timeout time.Duration) Suggester {
	if timeout < 5*time.Second {
		timeout = 60 * time.Second
	}
	return &openaiClient{
		apiKey:  apiKey,
		model:   model,
		baseURL: "https://api.openai.com/v1",
		client:  &http.Client{Timeout: timeout},
	}
}

func (c *openaiClient) SuggestMetadata(ctx context.Context, in Input) (Result, error) {
	start := time.Now()
	sys := `You help podcast catalog operators. Respond with a single JSON object only (no markdown) with keys:
summary: string, max 280 characters, neutral catalog-style description for operators.
operator_tags: array of 4-10 short lowercase slug strings (e.g. "technology", "interviews") for internal filtering, not duplicate of every iTunes category.
language: ISO 639-1 code for the show's primary language (best guess).
confidence: number between 0 and 1.

Base your answer only on the provided metadata; if unsure, lower confidence and keep tags generic.`

	userBits, _ := json.Marshal(in)
	user := fmt.Sprintf("Podcast metadata JSON:\n%s", userBits)

	body := map[string]any{
		"model": c.model,
		"messages": []map[string]string{
			{"role": "system", "content": sys},
			{"role": "user", "content": user},
		},
		"temperature":     0.3,
		"response_format": map[string]string{"type": "json_object"},
	}
	rawBody, err := json.Marshal(body)
	if err != nil {
		return Result{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(rawBody))
	if err != nil {
		return Result{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	res, err := c.client.Do(req)
	if err != nil {
		return Result{}, err
	}
	defer res.Body.Close()
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return Result{}, err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return Result{}, fmt.Errorf("openai http %d: %s", res.StatusCode, truncate(string(b), 500))
	}

	var cr completionsResponse
	if err := json.Unmarshal(b, &cr); err != nil {
		return Result{}, err
	}
	if len(cr.Choices) < 1 {
		return Result{}, errors.New("openai: empty choices")
	}
	content := strings.TrimSpace(cr.Choices[0].Message.Content)
	var ms domain.MetadataSuggestion
	if err := json.Unmarshal([]byte(content), &ms); err != nil {
		return Result{}, fmt.Errorf("openai: decode suggestion json: %w", err)
	}
	latencyMS := int(time.Since(start) / time.Millisecond)
	return Result{
		Suggestion: ms,
		LatencyMS:  latencyMS,
		Model:      c.model,
		Provider:   "openai",
	}, nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}

type completionsResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}
