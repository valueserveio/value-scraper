package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockScrapeData is a mock version of ScrapeData that sets generateSummary to false
func MockScrapeData(url string) (ScrapedData, error) {
	return ScrapeData(url, false)
}

func TestScrapeHandler(t *testing.T) {
	// Mock server to serve HTML content for URL
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`
            <!DOCTYPE html>
            <html>
            <head>
                <title>Test Title</title>
                <meta name="description" content="Test Description">
            </head>
            <body>
                <main>
                    <h1>Main Heading</h1>
                    <p>Some paragraph text.</p>
                </main>
            </body>
            </html>
        `))
	}))
	defer mockServer.Close()

	// Create a request payload
	requestPayload := requestBody{
		WebpageURL: mockServer.URL,
	}

	// Marshal the request payload to JSON
	jsonPayload, err := json.Marshal(requestPayload)
	assert.NoError(t, err)

	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodPost, "/scrape", bytes.NewBuffer(jsonPayload))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Create a handler function that uses the mock ScrapeData function
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		var requestPayload requestBody
		if err := json.NewDecoder(r.Body).Decode(&requestPayload); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		webpageData, err := MockScrapeData(requestPayload.WebpageURL)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error scraping URL: %v", err), http.StatusInternalServerError)
			return
		}

		response := responseBody{
			WebpageData: webpageData,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
		}
	})

	// Serve the HTTP request
	handler.ServeHTTP(rr, req)

	// Check the status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Decode the response body
	var response responseBody
	err = json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)

	// Verify the scraped data
	assert.Equal(t, mockServer.URL, response.WebpageData.URL)
	assert.Equal(t, "Test Title", response.WebpageData.Title)
	assert.Equal(t, "Test Description", response.WebpageData.Description)
}
