package main

import (
	"fmt"
	"log"
	"strings"

	pkg "golang-web-scraper/pkg"
)

func main() {
	config := pkg.DefaultConfig()
	content, err := pkg.ScrapeURL("https://medium.com/data-science-at-microsoft/how-large-language-models-work-91c362f5b78f", config)
	if err != nil {
		log.Fatal(err)
	}

	singleString := createSingleString(*content)

	fmt.Println(singleString)

	embeddingsConfig := pkg.DefaultEmbeddingConfig()
	chunk := singleString
	embeddings, err := pkg.GenerateEmbeddings(chunk, embeddingsConfig)
	if err != nil {
		fmt.Printf("Error generating embeddings: %v\n", err)
		return
	}
	fmt.Println(embeddings)
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
