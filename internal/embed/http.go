package embed

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPEmbedder calls a remote HTTP embedding service
type HTTPEmbedder struct {
	endpoint  string
	dimension int
	modelName string
	client    *http.Client
}

// EmbedRequest is the JSON request format
type EmbedRequest struct {
	Text string `json:"text"`
}

// EmbedResponse is the JSON response format
type EmbedResponse struct {
	Vector []float32 `json:"vector"`
}

// NewHTTPEmbedder creates an embedder that calls an HTTP endpoint
func NewHTTPEmbedder(endpoint string, dimension int, modelName string) *HTTPEmbedder {
	return &HTTPEmbedder{
		endpoint:  endpoint,
		dimension: dimension,
		modelName: modelName,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Embed calls the HTTP endpoint to embed text
func (h *HTTPEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	reqBody, err := json.Marshal(EmbedRequest{Text: text})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", h.endpoint, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("embedding service error %d: %s", resp.StatusCode, string(body))
	}

	var embedResp EmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&embedResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if len(embedResp.Vector) != h.dimension {
		return nil, fmt.Errorf("dimension mismatch: got %d, expected %d",
			len(embedResp.Vector), h.dimension)
	}

	return embedResp.Vector, nil
}

// Dimension returns the embedding dimension
func (h *HTTPEmbedder) Dimension() int {
	return h.dimension
}

// ModelName returns the model identifier
func (h *HTTPEmbedder) ModelName() string {
	return h.modelName
}
