package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScrapeData(t *testing.T) {
	// Mock server to serve HTML content
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

	// Call the ScrapeData function
	data, err := ScrapeData(mockServer.URL, false)
	assert.NoError(t, err)

	// Verify the scraped data
	assert.Equal(t, mockServer.URL, data.URL)
	assert.Equal(t, "Test Title", data.Title)
	assert.Equal(t, "Test Description", data.Description)
	// assert.Contains(t, data.AllText, "Main Heading")
	// assert.Contains(t, data.AllText, "Some paragraph text.")
}
