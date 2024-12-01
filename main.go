package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func main() {
	// Handle the /scrape endpoint
	http.HandleFunc("/scrape", scrapeHandler)

	// Start the server on port 3000
	log.Println("Server is starting on port 3000...")
	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// Handler function to process the scrape request
func scrapeHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the URL parameter from the query string
	url := r.URL.Query().Get("url")
	if url == "" {
		http.Error(w, "Missing 'url' query parameter", http.StatusBadRequest)
		return
	}

	// Call the scraper function with the provided URL
	companyInfo, err := scraper(url)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error scraping data: %v", err), http.StatusInternalServerError)
		return
	}

	for i := range companyInfo {
		s := ScrapedDataAI{}
		summarizedText, err := s.Summarize(companyInfo[i])
		if err != nil {
			http.Error(w, fmt.Sprintf("Error summarizing text: %v", err), http.StatusInternalServerError)
			return
		}
		fmt.Printf("Summary for range %v of url(%v): %v", i, url, summarizedText)
		companyInfo[i].Summary = string(summarizedText)
	}

	// Set the response header for JSON content
	w.Header().Set("Content-Type", "application/json")

	// Write the company info as JSON response
	if err := json.NewEncoder(w).Encode(companyInfo); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
	}

	fmt.Println("Done scraping:", url)
}
