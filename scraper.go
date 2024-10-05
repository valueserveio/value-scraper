package main

import (
	"encoding/json"
	"log"
	"regexp"
	"strings"

	"github.com/gocolly/colly/v2"
)

type ScrapedData struct {
	URL         string `json:"url,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Summary     string `json:"summary,omitempty"`
	// AllText     string `json:"allText,omitempty"`
}

// TODO: Update tests

func ScrapeData(url string) ([]byte, error) {
	data := ScrapedData{URL: url}
	var allText strings.Builder
	uniqueText := make(map[string]struct{})
	c := colly.NewCollector()

	data.URL = url

	// Define what to scrape
	c.OnHTML("title", func(e *colly.HTMLElement) {
		data.Title = e.Text
	})

	c.OnHTML("meta[name=description]", func(e *colly.HTMLElement) {
		data.Description = e.Attr("content")
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

	s := ScrapedDataAI{}

	summarizedText, err := s.Summarize(allText.String())
	if err != nil {
		log.Println("Error summarizing text:", err)
		return nil, err
	}

	s.Summary = summarizedText
	// fmt.Println("Summary:", s.Summary)

	data.Summary = string(summarizedText)
	//data.AllText = allText.String()

	// Convert struct to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return jsonData, nil

	// data["allText"] = allText.String()
}
