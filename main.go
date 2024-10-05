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
	CustomerURL string `json:"customerUrl"`
	ProductURL  string `json:"productUrl"`
}

type responseBody struct {
	CustomerData ScrapedData `json:"customerData"`
	ProductData  ScrapedData `json:"productData"`
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

	customerData, err := ScrapeData(requestPayload.CustomerURL)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error scraping customer URL: %v", err), http.StatusInternalServerError)
		return
	}

	productData, err := ScrapeData(requestPayload.ProductURL)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error scraping product URL: %v", err), http.StatusInternalServerError)
		return
	}

	response := responseBody{
		CustomerData: customerData,
		ProductData:  productData,
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
	err := godotenv.Load(".env.local")
	if err != nil {
		log.Fatal("Error loading .env.local file")
	}

	http.HandleFunc("/scrape", scrapeHandler)
	log.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
