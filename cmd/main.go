package main

import (
	"fmt"
	"log"
	"strings"

	pkg "golang-web-scraper/pkg"
)

func main() {
	config := pkg.DefaultConfig()

	urls := []string{
		"https://medium.com/data-science-at-microsoft/how-large-language-models-work-91c362f5b78f",
		"https://www.reddit.com/r/golang/comments/1bdp1ku/hugot_hugginface_transformer_pipelines_for_golang/",
		"https://sbert.net",
		"https://medium.com/@denhox/sharing-data-between-microservices-fe7fb9471208",
		"https://medium.com/data-science-collective/the-open-source-stack-for-ai-agents-8ab900e33676",
	}

	results := pkg.ScrapeURLsConcurrently(urls, config, 5)
	for _, content := range results {
		if content != nil {
			fmt.Println(createSingleString(*content))
			fmt.Println("Results Generated")
		} else {
			log.Println("Failed to scrape a URL")
		}
	}

	// var wg sync.WaitGroup

	// for _, content := range results {
	// 	if content != nil {
	// 		wg.Add(1)

	// 		go func(o pkg.ScrapedContent) {
	// 			defer wg.Done()
	// 			embeddings, err := pkg.GenerateEmbeddings(o)
	// 			if err != nil {
	// 				log.Fatalf("Failed to generate embedding: %v", err)
	// 			}

	// 			for _, embedding := range embeddings {
	// 				// fmt.Printf("Embedding %d: %v\n", i, embedding)
	// 				if embedding != nil {

	// 				} else {
	// 					log.Println("Failed to generate embeddings")
	// 				}

	// 			}
	// 		}(*content)

	// 		wg.Wait()
	// 		fmt.Println("Embeddings Done")

	// 	}
	// }

}

func createSingleString(content pkg.ScrapedContent) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Title: %s\n", content.Title))
	sb.WriteString(fmt.Sprintf("Description: %s\n", content.Description))

	sb.WriteString("Headers:\n")
	for _, header := range content.Headers {
		sb.WriteString(fmt.Sprintf("  - %s\n", header))
	}

	sb.WriteString("Paragraphs:\n")
	for _, paragraph := range content.Paragraphs {
		sb.WriteString(fmt.Sprintf("  - %s\n", paragraph))
	}

	sb.WriteString("Links:\n")
	for _, link := range content.Links {
		sb.WriteString(fmt.Sprintf("  - %s\n", link))
	}

	return sb.String()
}
