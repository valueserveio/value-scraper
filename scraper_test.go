package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlerIntegration(t *testing.T) {
	// server for testing
	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	// data for the POST request
	requestData := map[string]string{
		"customer_url": "http://berops.com",
		"product_url":  "http://darktrace.com",
	}
	requestBody, err := json.Marshal(requestData)
	if err != nil {
		t.Fatalf("Failed to marshal request data: %v", err)
	}

	// POST request
	req, err := http.NewRequest(http.MethodPost, ts.URL, bytes.NewBuffer(requestBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// response check
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Parsing response body
	var responseData struct {
		CustomerData map[string]string `json:"customer_data"`
		ProductData  map[string]string `json:"product_data"`
	}
	err = json.NewDecoder(resp.Body).Decode(&responseData)
	if err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	// Additional checks for responseData
	if len(responseData.CustomerData) > 0 {
		//t.Logf("Customer Data: %v\n", responseData.CustomerData)
	} else {
		t.Error("Expected customer data, got empty map")
	}

	if len(responseData.ProductData) > 0 {
		//t.Logf("Product Data: %v\n", responseData.ProductData)
	} else {
		t.Error("Expected product data, got empty map")
	}
}
