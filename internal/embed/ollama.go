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

// OllamaEmbedder uses Ollama's embedding API
type OllamaEmbedder struct {
	endpoint  string
	model     string
	dimension int
	client    *http.Client
}

// OllamaEmbedRequest is the request format for Ollama
type OllamaEmbedRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

// OllamaEmbedResponse is the response format from Ollama
type OllamaEmbedResponse struct {
	Embedding []float32 `json:"embedding"`
}

// NewOllamaEmbedder creates an embedder using Ollama
// Default endpoint: http://localhost:11434
// Recommended models: nomic-embed-text, all-minilm
func NewOllamaEmbedder(endpoint, model string, dimension int) *OllamaEmbedder {
	if endpoint == "" {
		endpoint = "http://localhost:11434"
	}
	return &OllamaEmbedder{
		endpoint:  endpoint,
		model:     model,
		dimension: dimension,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Embed calls Ollama to embed text
func (o *OllamaEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	reqBody, err := json.Marshal(OllamaEmbedRequest{
		Model:  o.model,
		Prompt: text,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/embeddings", o.endpoint)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama error %d: %s", resp.StatusCode, string(body))
	}

	var embedResp OllamaEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&embedResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if len(embedResp.Embedding) == 0 {
		return nil, fmt.Errorf("empty embedding returned")
	}

	// Update dimension if not set
	if o.dimension == 0 {
		o.dimension = len(embedResp.Embedding)
	}

	return embedResp.Embedding, nil
}

// Dimension returns the embedding dimension
func (o *OllamaEmbedder) Dimension() int {
	return o.dimension
}

// ModelName returns the model identifier
func (o *OllamaEmbedder) ModelName() string {
	return fmt.Sprintf("ollama-%s", o.model)
}

// Ping checks if Ollama is available
func (o *OllamaEmbedder) Ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", o.endpoint+"/api/tags", nil)
	if err != nil {
		return err
	}

	resp, err := o.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama not available: status %d", resp.StatusCode)
	}

	return nil
}
