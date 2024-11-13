package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/joho/godotenv"
)

type requestBody struct {
	WebpageURL string `json:"url"`
}

type responseBody struct {
	WebpageData ScrapedData `json:"WebpageData"`
}

func scrapeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var requestPayload requestBody
	if err := json.NewDecoder(r.Body).Decode(&requestPayload); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	WebpageData, err := ScrapeData(requestPayload.WebpageURL, true)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error scraping URL: %v", err), http.StatusInternalServerError)
		return
	}

	response := responseBody{
		WebpageData: WebpageData,
	}

	// Marshal the response struct to JSON
	jsonData, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Error marshaling JSON response", http.StatusInternalServerError)
		return
	}

	// Format the JSON output
	var formattedJSON bytes.Buffer
	if err := json.Indent(&formattedJSON, jsonData, "", "  "); err != nil {
		http.Error(w, "Error formatting JSON response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(formattedJSON.Bytes())
}

func main() {
	// Load .env.local file if it exists
	if err := godotenv.Load(".env.local"); err != nil {
		log.Printf("No .env.local file found: %v", err)
	}

	http.HandleFunc("/scrape", scrapeHandler)
	log.Println("Server started at :3000")
	log.Fatal(http.ListenAndServe(":3000", nil))
}
