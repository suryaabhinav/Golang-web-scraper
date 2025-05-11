package pkg

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
)

// ScrapedContent represents the structured content from a webpage
type ScrapedContent struct {
	Title       string
	Description string
	Headers     []string
	Paragraphs  []string
	Links       []string
}

// ScrapeConfig holds configuration for the scraper
type ScrapeConfig struct {
	Timeout   time.Duration
	MaxDepth  int
	UserAgent string
}

// DefaultConfig returns a default scraping configuration
func DefaultConfig() *ScrapeConfig {
	return &ScrapeConfig{
		Timeout:   30 * time.Second,
		MaxDepth:  2,
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
	}
}

// ScrapeURL scrapes content from the given URL and extracts meaningful information
func ScrapeURL(url string, config *ScrapeConfig) (*ScrapedContent, error) {
	if config == nil {
		config = DefaultConfig()
	}

	content := &ScrapedContent{}

	// Initialize collector with configuration
	c := colly.NewCollector(
		colly.MaxDepth(config.MaxDepth),
		colly.IgnoreRobotsTxt(),
	)

	// Set timeout
	c.WithTransport(&http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           (&net.Dialer{Timeout: config.Timeout}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	})

	// Add extensions
	extensions.RandomUserAgent(c)
	extensions.Referer(c)

	// Set custom headers
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("User-Agent", config.UserAgent)
	})

	// Extract title
	c.OnHTML("title", func(e *colly.HTMLElement) {
		content.Title = strings.TrimSpace(e.Text)
	})

	// Extract meta description
	c.OnHTML("meta[name=description]", func(e *colly.HTMLElement) {
		content.Description = strings.TrimSpace(e.Attr("content"))
	})

	// Extract headers (h1-h3 for relevance)
	c.OnHTML("h1, h2, h3", func(e *colly.HTMLElement) {
		text := strings.TrimSpace(e.Text)
		if text != "" {
			content.Headers = append(content.Headers, text)
		}
	})

	// Extract paragraphs
	c.OnHTML("p", func(e *colly.HTMLElement) {
		text := cleanText(e.Text)
		if text != "" {
			content.Paragraphs = append(content.Paragraphs, text)
		}
	})

	// Error handling
	c.OnError(func(r *colly.Response, err error) {
		log.Printf("Error scraping %s: %v", r.Request.URL, err)
	})

	// Start scraping
	if err := c.Visit(url); err != nil {
		return nil, fmt.Errorf("failed to scrape URL %s: %w", url, err)
	}

	return content, nil
}

// cleanText processes text by removing unwanted patterns and normalizing whitespace
func cleanText(text string) string {
	// Remove extra whitespace
	text = regexp.MustCompile(`\s+`).ReplaceAllString(strings.TrimSpace(text), " ")

	// Clean common HTML entities
	text = strings.NewReplacer(
		"&nbsp;", " ",
		"&amp;", "&",
		"&lt;", "<",
		"&gt;", ">",
		"&quot;", "\"",
		"&apos;", "'",
	).Replace(text)

	return text
}

// ScrapeURLsConcurrently scrapes multiple URLs concurrently and returns their content
func ScrapeURLsConcurrently(urls []string, config *ScrapeConfig, numWorkers int) []*ScrapedContent {
	if config == nil {
		config = DefaultConfig()
	}

	urlChan := make(chan string, len(urls))
	resultChan := make(chan *ScrapedContent, len(urls))
	var wg sync.WaitGroup

	// Worker function
	worker := func() {
		defer wg.Done()
		for url := range urlChan {
			content, err := ScrapeURL(url, config)
			if err != nil {
				log.Printf("Failed to scrape URL %s: %v", url, err)
				continue
			}
			resultChan <- content
		}
	}

	// Start workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker()
	}

	// Send URLs to the channel
	for _, url := range urls {
		urlChan <- url
	}
	close(urlChan)

	// Wait for all workers to finish
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	var results []*ScrapedContent
	for content := range resultChan {
		results = append(results, content)
	}

	return results
}
