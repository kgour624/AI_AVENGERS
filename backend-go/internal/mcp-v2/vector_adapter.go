package mcpv2

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// VectorAdapter calls Python sidecar bge-base-en-v1.5 — 768D (RULE 8-F: reuse ml-sidecar).
// Keeps 1 Uvicorn worker + bounded ThreadPool (sidecar side), here just HTTP with timeout.
type VectorAdapter struct {
	baseURL string
	client  *http.Client
}

func NewVectorAdapter(baseURL string) *VectorAdapter {
	if baseURL == "" {
		baseURL = "http://ml-sidecar:8001"
	}
	return &VectorAdapter{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (v *VectorAdapter) Embed(ctx context.Context, text string) ([]float32, error) {
	if text == "" {
		return nil, fmt.Errorf("text required")
	}
	body, _ := json.Marshal(map[string]string{"text": text})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, v.baseURL+"/embed", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := v.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sidecar embed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sidecar status %d", resp.StatusCode)
	}
	var out struct {
		Embedding []float32 `json:"embedding"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if len(out.Embedding) != 768 {
		return nil, fmt.Errorf("unexpected embedding dim %d", len(out.Embedding))
	}
	return out.Embedding, nil
}