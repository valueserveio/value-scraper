package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

var req struct {
	URL string `json:"url"`
}

// Handler function to process the scrape request
func scrapeHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, "Missing 'url' in request body", http.StatusBadRequest)
		return
	}

	// Call the scraper function with the provided URL
	companyInfo, err := scraper(req.URL)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error scraping data: %v", err), http.StatusInternalServerError)
		return
	}

	// Summarize the text for each company info
	for i := range companyInfo {
		s := ScrapedDataAI{}
		summarizedText, err := s.Summarize(companyInfo[i])
		if err != nil {
			http.Error(w, fmt.Sprintf("Error summarizing text: %v", err), http.StatusInternalServerError)
			return
		}
		fmt.Printf("Summary for range %v of url(%v): %v", i, req.URL, summarizedText)
		companyInfo[i].Summary = string(summarizedText)
	}

	// Set the response header for JSON content
	w.Header().Set("Content-Type", "application/json")

	// Write the company info as JSON response
	if err := json.NewEncoder(w).Encode(companyInfo); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
	}

	fmt.Println("Done scraping:", req.URL)
}

func main() {
	// Handle the /scrape endpoint
	http.HandleFunc("/scrape", scrapeHandler)

	// Start the server on port 3000
	log.Println("Server is starting on port 3000...")
	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
