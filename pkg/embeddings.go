package pkg

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/milosgajdos/go-embeddings/ollama"
)

// // ScrapedContent represents the structure of scraped web content
// type ScrapedContent struct {
// 	Title       string
// 	Description string
// 	Headers     []string
// 	Paragraphs  []string
// 	Links       []string
// }

var (
	model string
	debug bool
)

func init() {
	flag.StringVar(&model, "model", "llama3", "Ollama model name")
	flag.BoolVar(&debug, "debug", false, "Enable debug logging")
}

// chunkContent breaks down the scraped content into manageable chunks
func chunkContent(content ScrapedContent, maxChunkSize int) []string {
	chunks := []string{}

	// Add title and description as initial chunks
	if content.Title != "" {
		chunks = append(chunks, content.Title)
	}
	if content.Description != "" {
		chunks = append(chunks, content.Description)
	}

	// Add headers
	chunks = append(chunks, content.Headers...)

	// Break paragraphs into chunks
	for _, paragraph := range content.Paragraphs {
		// If paragraph is longer than maxChunkSize, split it
		if len(paragraph) > maxChunkSize {
			words := strings.Split(paragraph, " ")
			currentChunk := ""
			for _, word := range words {
				if len(currentChunk+" "+word) > maxChunkSize {
					chunks = append(chunks, strings.TrimSpace(currentChunk))
					currentChunk = word
				} else {
					currentChunk += " " + word
				}
			}
			if currentChunk != "" {
				chunks = append(chunks, strings.TrimSpace(currentChunk))
			}
		} else {
			chunks = append(chunks, paragraph)
		}
	}

	return chunks
}

// GenerateEmbeddings generates embeddings for scraped content concurrently
func GenerateEmbeddings(content ScrapedContent) ([][]float32, error) {
	flag.Parse()

	if model == "" {
		return nil, fmt.Errorf("missing Ollama model")
	}

	// Chunk the content
	chunks := chunkContent(content, 1024) // 1024 is an example max chunk size

	// Prepare for concurrent processing
	var wg sync.WaitGroup
	var mu sync.Mutex
	allEmbeddings := make([][]float32, len(chunks))
	errorChan := make(chan error, len(chunks))

	// Create Ollama client
	c := ollama.NewClient()

	// Process chunks concurrently
	for i, chunk := range chunks {
		wg.Add(1)
		go func(index int, text string) {
			defer wg.Done()

			embReq := &ollama.EmbeddingRequest{
				Prompt: text,
				Model:  model,
			}

			embs, err := c.Embed(context.Background(), embReq)
			if err != nil {
				if debug {
					log.Printf("Error generating embedding for chunk %d: %v", index, err)
				}
				errorChan <- err
				return
			}

			mu.Lock()
			flattenedEmbs := []float32{}
			for _, emb := range embs {
				for _, v := range emb.Vector {
					flattenedEmbs = append(flattenedEmbs, float32(v))
				}
			}
			allEmbeddings[index] = flattenedEmbs
			mu.Unlock()
		}(i, chunk)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Check for any errors
	select {
	case err := <-errorChan:
		return nil, err
	default:
		// No errors
	}

	// Filter out any nil embeddings
	var filteredEmbeddings [][]float32
	for _, emb := range allEmbeddings {
		if emb != nil {
			filteredEmbeddings = append(filteredEmbeddings, emb)
		}
	}

	if debug {
		fmt.Printf("Generated %d embeddings from %d chunks\n", len(filteredEmbeddings), len(chunks))
	}

	return filteredEmbeddings, nil
}
