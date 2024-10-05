package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/joho/godotenv"
)

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

	var requestData struct {
		CustomerURL string `json:"customer_url"`
		ProductURL  string `json:"product_url"`
	}
	err = json.Unmarshal(body, &requestData)
	if err != nil {
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}

	customerData, err := ScrapeData(requestData.CustomerURL)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to scrape customer URL: %v", err), http.StatusInternalServerError)
		return
	}

	productData, err := ScrapeData(requestData.ProductURL)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to scrape product URL: %v", err), http.StatusInternalServerError)
		return
	}

	responseData := struct {
		CustomerData []byte `json:"customer_data"`
		ProductData  []byte `json:"product_data"`
	}{
		CustomerData: customerData,
		ProductData:  productData,
	}

	// TODO: response needs type conversion | example response output: {"customer_data":"eyJ1cmwiOiJodHRwOi8vYmVyb3BzLmNvbSIsI......

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseData)
}

func main() {
	err := godotenv.Load(".env.local")
	if err != nil {
		log.Fatal("Error loading .env.local file")
	}

	http.HandleFunc("/scrape", handler)
	log.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
