package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// EmbeddingRequest represents the input payload for generating embeddings
type EmbeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

// EmbeddingResponse represents the output from the embedding generation
type EmbeddingResponse struct {
	Embeddings [][]float64 `json:"embeddings"`
	Error      string      `json:"error,omitempty"`
}

// EmbeddingConfig holds the configuration for the embedding generation process
type EmbeddingConfig struct {
	BaseURL   string
	ModelName string
	Timeout   time.Duration
}

// DefaultEmbeddingConfig returns the default configuration for the embedding service
func DefaultEmbeddingConfig() *EmbeddingConfig {
	return &EmbeddingConfig{
		BaseURL:   "http://localhost:11434/api/embeddings",
		ModelName: "llama3",
		Timeout:   30 * time.Second,
	}
}

// GenerateEmbeddings sends a REST API request to generate embeddings for the given text chunk
func GenerateEmbeddings(chunk string, config *EmbeddingConfig) ([][]float64, error) {
	if config == nil {
		config = DefaultEmbeddingConfig()
	}

	// Create the request payload
	payload := EmbeddingRequest{
		Model:  config.ModelName,
		Prompt: chunk,
	}

	requestBody, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal embedding request: %w", err)
	}

	// Create an HTTP client
	client := &http.Client{
		Timeout: config.Timeout,
	}

	// Make the POST request
	req, err := http.NewRequest("POST", config.BaseURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create embedding request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("embedding request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read and parse the response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read embedding response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embedding service returned status %d: %s", resp.StatusCode, string(responseBody))
	}

	var embeddingResponse EmbeddingResponse
	if err := json.Unmarshal(responseBody, &embeddingResponse); err != nil {
		return nil, fmt.Errorf("failed to parse embedding response: %w", err)
	}

	if embeddingResponse.Error != "" {
		return nil, fmt.Errorf("embedding service error: %s", embeddingResponse.Error)
	}

	return embeddingResponse.Embeddings, nil
}
