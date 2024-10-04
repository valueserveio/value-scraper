package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/gocolly/colly/v2"
)

// ScrapeData scrapes valuable data from the given URL
func ScrapeData(url string) (map[string]string, error) {
	data := make(map[string]string)
	var allText strings.Builder
	uniqueText := make(map[string]struct{})
	c := colly.NewCollector()

	data["URL"] = url

	// Define what to scrape
	c.OnHTML("title", func(e *colly.HTMLElement) {
		data["title"] = e.Text
	})

	c.OnHTML("meta[name=description]", func(e *colly.HTMLElement) {
		data["description"] = e.Attr("content")
	})

	// // Capture all text (replaced with other solution)
	// c.OnHTML("body", func(e *colly.HTMLElement) {
	// 	allText.WriteString(e.Text)
	// })

	// Detect JSON-like patterns
	jsonLikePattern := regexp.MustCompile(`\{.*\}|\[.*\]`)

	// Capture text from specific elements, excluding repetitive or non-informative sections
	c.OnHTML("main, article, section, div.content, div.main-content", func(e *colly.HTMLElement) {
		e.ForEach("p, h1, h2, h3, h4, h5, h6, li, span, div, a, td, th, blockquote, pre, code", func(_ int, el *colly.HTMLElement) {
			text := strings.TrimSpace(el.Text)
			if _, exists := uniqueText[text]; !exists && text != "" && !jsonLikePattern.MatchString(text) && !strings.Contains(text, "iframe") {
				uniqueText[text] = struct{}{}
				if !strings.Contains(allText.String(), text) {
					allText.WriteString(text)
				}
			}
		})
	})

	// Handle errors
	c.OnError(func(_ *colly.Response, err error) {
		log.Println("Something went wrong:", err)
	})

	// Visit the URL
	err := c.Visit(url)
	if err != nil {
		return nil, err
	}

	// // Remove duplicates from the collected text (replaced with other solution)
	// allTextSlice := strings.Split(allText.String(), "\n")
	// uniqueAllText := removeDuplicates(allTextSlice)
	// data["allText"] = strings.Join(uniqueAllText, "\n")

	s := ScrapedData{}

	s.OriginalText = allText.String()

	summarizedText, err := s.Summarize(allText.String())
	if err != nil {
		log.Println("Error summarizing text:", err)
		return nil, err
	}

	s.Summary = summarizedText
	fmt.Println("Summary:", s.Summary)

	data["allText"] = allText.String()

	return data, nil
}
