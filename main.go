package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/gocolly/colly/v2"
)

// // Function to remove duplicates from a slice of strings (replaced with other solution)
// func removeDuplicates(elements []string) []string {
// 	encountered := map[string]bool{}
// 	result := []string{}

// 	for v := range elements {
// 		if encountered[elements[v]] {
// 			// Do not add duplicate.
// 		} else {
// 			// Record this element as an encountered element.
// 			encountered[elements[v]] = true
// 			// Append to result slice.
// 			result = append(result, elements[v])
// 		}
// 	}
// 	return result
// }

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
					allText.WriteString(text + "\n")
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

	data["allText"] = allText.String()

	return data, nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var requestData map[string]string
	err = json.Unmarshal(body, &requestData)
	if err != nil {
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}

	url, exists := requestData["url"]
	if !exists {
		http.Error(w, "URL not provided", http.StatusBadRequest)
		return
	}

	data, err := ScrapeData(url)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to scrape URL: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func main() {
	http.HandleFunc("/scrape", handler)
	log.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// // for local testing
// func main() {
// 	// Define the URL to scrape
// 	url := "https://www.darktrace.com"

// 	// Scrape the data
// 	data, err := ScrapeData(url)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	// // Marshal the data map into a JSON-formatted string
// 	// jsonData, err := json.MarshalIndent(data, "", "  ")
// 	// if err != nil {
// 	// 	log.Fatal(err)
// 	// }

// 	// // Print the JSON-formatted string
// 	// fmt.Println(string(jsonData))

// 	// Print the scraped data
// 	fmt.Println("URL:", url)
// 	fmt.Println("Title:", data["title"])
// 	fmt.Println("Description:", data["description"])
// 	fmt.Println("All Text:", data["allText"])
// }
