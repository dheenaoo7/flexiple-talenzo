// Package llm wraps the Gemini API (google.golang.org/genai) for the three
// structured calls the search pipeline needs: drafting filters+rubric from a
// query, scoring/explaining candidates against a rubric, and evolving
// filters+rubric from recruiter feedback.
package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"google.golang.org/genai"
)

const callTimeout = 45 * time.Second

// Error is a typed LLM failure surfaced to the SSE stream so the frontend can
// show a retryable error state without losing recruiter input.
type Error struct {
	Stage     string
	Message   string
	Retryable bool
	Cause     error
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Stage, e.Message)
}

func (e *Error) Unwrap() error { return e.Cause }

type Client struct {
	raw   *genai.Client
	model string
}

func New(ctx context.Context, apiKey, model string) (*Client, error) {
	c, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("creating gemini client: %w", err)
	}
	return &Client{raw: c, model: model}, nil
}

func floatPtr(f float32) *float32 { return &f }

func int32Ptr(i int32) *int32 { return &i }

// generateJSON calls Gemini with a JSON response schema and unmarshals the
// result into out. It retries once on a transient network error or malformed
// JSON before giving up.
func generateJSON(ctx context.Context, c *Client, stage, systemInstruction, userPrompt string, schema *genai.Schema, out any) error {
	cfg := &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{Parts: []*genai.Part{{Text: systemInstruction}}},
		ResponseMIMEType:  "application/json",
		ResponseSchema:    schema,
		Temperature:       floatPtr(0.3),
		// Structured extraction/scoring doesn't need chain-of-thought; disabling
		// it keeps latency well under Gemini's own gateway timeout.
		ThinkingConfig: &genai.ThinkingConfig{ThinkingBudget: int32Ptr(0)},
	}

	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		text, err := callOnce(ctx, c, userPrompt, cfg)
		if err != nil {
			lastErr = err
			continue
		}
		if err := json.Unmarshal([]byte(text), out); err != nil {
			lastErr = fmt.Errorf("parsing model response as JSON: %w (raw: %.200s)", err, text)
			continue
		}
		return nil
	}

	return &Error{Stage: stage, Message: friendlyMessage(lastErr), Retryable: isRetryable(lastErr), Cause: lastErr}
}

func callOnce(ctx context.Context, c *Client, userPrompt string, cfg *genai.GenerateContentConfig) (string, error) {
	callCtx, cancel := context.WithTimeout(ctx, callTimeout)
	defer cancel()

	resp, err := c.raw.Models.GenerateContent(callCtx, c.model, genai.Text(userPrompt), cfg)
	if err != nil {
		return "", fmt.Errorf("calling gemini: %w", err)
	}
	text := strings.TrimSpace(resp.Text())
	if text == "" {
		return "", fmt.Errorf("model returned an empty response (it may have been blocked)")
	}
	return text, nil
}

func friendlyMessage(err error) string {
	if err == nil {
		return "the model failed for an unknown reason"
	}
	msg := err.Error()
	switch {
	case strings.Contains(msg, "empty response"):
		return "the model declined to respond — try rephrasing your query"
	case strings.Contains(msg, "parsing model response"):
		return "the model returned an unexpected format — please retry"
	case strings.Contains(msg, "RESOURCE_EXHAUSTED") || strings.Contains(msg, "429"):
		return "the model's rate limit/quota was hit — wait a bit before retrying"
	case strings.Contains(msg, "UNAVAILABLE") || strings.Contains(msg, "DEADLINE_EXCEEDED"):
		return "the model is temporarily unavailable — please retry"
	default:
		return "couldn't reach the model — check your connection and retry"
	}
}

// isRetryable is false for quota errors: retrying immediately just burns
// another request against the same exhausted quota.
func isRetryable(err error) bool {
	if err == nil {
		return true
	}
	msg := err.Error()
	return !strings.Contains(msg, "RESOURCE_EXHAUSTED") && !strings.Contains(msg, "429")
}
