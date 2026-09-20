package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"talent-search-rubric-arm/backend/internal/llm"
)

type sseEvent struct {
	Type      string `json:"type"`
	Stage     string `json:"stage,omitempty"`
	Message   string `json:"message,omitempty"`
	Retryable bool   `json:"retryable,omitempty"`
	Payload   any    `json:"payload,omitempty"`
}

// sseWriter streams JSON events as text/event-stream frames, flushing after
// each one so the client sees real, incremental "thinking" progress.
type sseWriter struct {
	w       http.ResponseWriter
	flusher http.Flusher
}

func newSSEWriter(w http.ResponseWriter) (*sseWriter, bool) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, false
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()
	return &sseWriter{w: w, flusher: flusher}, true
}

func (s *sseWriter) send(ev sseEvent) {
	payload, err := json.Marshal(ev)
	if err != nil {
		log.Printf("sse: failed to marshal event: %v", err)
		return
	}
	if _, err := s.w.Write([]byte("data: ")); err != nil {
		return
	}
	if _, err := s.w.Write(payload); err != nil {
		return
	}
	if _, err := s.w.Write([]byte("\n\n")); err != nil {
		return
	}
	s.flusher.Flush()
}

func (s *sseWriter) stage(stage, message string) {
	s.send(sseEvent{Type: "stage", Stage: stage, Message: message})
}

func (s *sseWriter) result(eventType string, payload any) {
	s.send(sseEvent{Type: eventType, Payload: payload})
}

func (s *sseWriter) error(err error) {
	var llmErr *llm.Error
	if errors.As(err, &llmErr) {
		s.send(sseEvent{Type: "error", Stage: llmErr.Stage, Message: llmErr.Message, Retryable: llmErr.Retryable})
		return
	}
	s.send(sseEvent{Type: "error", Stage: "pipeline", Message: err.Error(), Retryable: false})
}
