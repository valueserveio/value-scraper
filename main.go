package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

type SERPResults struct {
	OrganicSearchResults []OrganicSearchResult `json:"organic"`
}

type OrganicSearchResult struct {
	Link        string `json:"link,omitempty"`
	DisplayLink string `json:"display_link,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Rank        int    `json:"rank,omitempty"`
	GlobalRank  int    `json:"global_rank,omitempty"`
}

type SnapshotStatus struct {
	SnapShotID string `json:"snapshot_id,omitempty"`
	Status     string `json:"status,omitempty"`
	Message    string `json:"message,omitempty"`
}

// Coming from crunchbase
type CompanyInfo struct {
	FullDescription string           `json:"full_description,omitempty"`
	Products        []ProductDetails `json:"products_and_services,omitempty"`
	EmployeeCount   string           `json:"num_employees"`
}

type ProductDetails struct {
	ProductName        string `json:"product_name,omitempty"`
	ProductDescription string `json:"product_description,omitempty"`
}

type URLPayload struct {
	URL string `json:"url"`
}

// Company profile
// Crunchbase, Owler, Zoominfo
// Product Descriptions
// Company Description

func main() {

	// Load .env.local file if it exists
	if err := godotenv.Load(".env.local"); err != nil {
		fmt.Printf("No .env.local file found: %v", err)
	}

	company_website := "https://www.ups.com/us/en/home"
	//customer_website := "https://topgolf.com"

	// Set up the proxy URL
	proxyURL, err := url.Parse(fmt.Sprintf("http://%v:%v@brd.superproxy.io:22225", os.Getenv("SERP_USERNAME"), os.Getenv("SERP_PASSWORD")))
	if err != nil {
		fmt.Println("Error parsing proxy URL:", err)
		return
	}

	company_request_url, err := url.Parse(fmt.Sprintf("https://www.google.com/search?q=%v+crunchbase&brd_json=1", company_website))
	if err != nil {
		fmt.Println("Error company search URL:", err)
		return
	}

	company_search_results, err := GetSERPData(company_request_url, proxyURL)
	if err != nil {
		fmt.Println("Error getting Company SERP data", err)
	}

	//fmt.Printf("Search Results: %+v\n", company_search_results)

	topSearch := []URLPayload{} 
	for i, v := range company_search_results.OrganicSearchResults {
		// Get top search
		if i > 0 {
			break
		}
		topSearch = append(topSearch, URLPayload{URL: v.Link})
	}

	fmt.Println("Topsearch result: ", topSearch)

	// Build Organization profile with Crunchbase

	// Fetch company info
	companyInfo, err := FetchCompanyInfo(topSearch)
	if err != nil {
		fmt.Println("Error fetching company profile:", err)
		return
	}

	// Print the result
	fmt.Printf("Company Info: %+v\n", companyInfo)

}

func GetSERPData(searchURL *url.URL, proxyURL *url.URL) (*SERPResults, error) {
	// Create an HTTP transport with proxy and skip SSL verification
	fmt.Println("Getting SERP data...")
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // Ignore SSL certificate errors
		},
	}

	// Create an HTTP client
	client := &http.Client{
		Transport: transport,
	}
	// Create the request
	req, err := http.NewRequest("GET", searchURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("Error creating request: %v", err)
	}

	// Perform the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error making request: %v", err)
	}
	defer resp.Body.Close()

	// Read and print the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading response: %v", err)
	}

	// Unmarshal the JSON response into the SERPResults struct
	var results SERPResults
	err = json.Unmarshal(body, &results)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshaling JSON: %v", err)
	}

	fmt.Printf("Search Results: %+v\n", results)

	return &results, nil

}

// Function to send request and get company info
func FetchCompanyInfo(payloadData []URLPayload) ([]CompanyInfo, error) {

	fmt.Println("Fetching company info...")
	// Marshal the payload into JSON
	payload, err := json.Marshal(payloadData)
	if err != nil {
		return []CompanyInfo{}, fmt.Errorf("error marshaling payload: %v", err)
	}

	// Define the URL
	url := "https://api.brightdata.com/datasets/v3/trigger?dataset_id=gd_l1vijqt9jfj7olije&include_errors=true"

	// Create the request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return []CompanyInfo{}, fmt.Errorf("error creating request: %v", err)
	}

	// Set headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", os.Getenv("BRIGHT_TOKEN")))
	req.Header.Set("Content-Type", "application/json")

	// Perform the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return []CompanyInfo{}, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []CompanyInfo{}, fmt.Errorf("error reading response body: %v", err)
	}

	var snapshotStatus SnapshotStatus
	err = json.Unmarshal(body, &snapshotStatus)
	if err != nil {
		return []CompanyInfo{}, fmt.Errorf("error unmarshaling snapshot response: %v", err)
	}

	// Recursively checks snapshot status until company info is ready
	snapShotStatus, snapshotErr := FetchSnapshotStatus(snapshotStatus.SnapShotID)
	if snapshotErr != nil {
		return []CompanyInfo{}, fmt.Errorf("error getting snapshot status: %v", err)

	}

	if snapShotStatus.Status == "complete" {
		// Replace the snapshot ID in the URL
		url := fmt.Sprintf("https://api.brightdata.com/datasets/v3/snapshot/%s?format=json", snapShotStatus.SnapShotID)

		// Get the Bearer token from the environment variable
		token := os.Getenv("BRIGHT_TOKEN")
		if token == "" {
			return []CompanyInfo{}, fmt.Errorf("environment variable BRIGHT_TOKEN is not set")
		}

		// Create the request
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return []CompanyInfo{}, fmt.Errorf("error creating request: %v", err)
		}

		// Set the Authorization header
		req.Header.Set("Authorization", "Bearer "+token)

		// Perform the request
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return []CompanyInfo{}, fmt.Errorf("error making request: %v", err)
		}
		defer resp.Body.Close()

		// Read the response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return []CompanyInfo{}, fmt.Errorf("error reading response body: %v", err)
		}

		//Unmarshal the response into CompanyInfo
		var companyInfo []CompanyInfo
		err = json.Unmarshal(body, &companyInfo)
		if err != nil {
			return []CompanyInfo{}, fmt.Errorf("error unmarshaling response: %v", err)
		}

		return companyInfo, nil

	}

	return []CompanyInfo{}, fmt.Errorf("No company information found.")

}

// Function to fetch snapshot status by ID
func FetchSnapshotStatus(snapshotID string) (SnapshotStatus, error) {
	time.Sleep(3 * time.Second)

	// Replace the snapshot ID in the URL
	url := fmt.Sprintf("https://api.brightdata.com/datasets/v3/snapshot/%s?format=json", snapshotID)

	// Get the Bearer token from the environment variable
	token := os.Getenv("BRIGHT_TOKEN")
	if token == "" {
		return SnapshotStatus{}, fmt.Errorf("environment variable BRIGHT_TOKEN is not set")
	}

	// Create the request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return SnapshotStatus{}, fmt.Errorf("error creating request: %v", err)
	}

	// Set the Authorization header
	req.Header.Set("Authorization", "Bearer "+token)

	// Perform the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return SnapshotStatus{}, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return SnapshotStatus{}, fmt.Errorf("error reading response body: %v", err)
	}

	// Unmarshal the response into SnapshotStatus
	var snapshotStatus SnapshotStatus
	err = json.Unmarshal(body, &snapshotStatus)
	if err != nil {
		return SnapshotStatus{}, fmt.Errorf("error unmarshaling JSON: %v", err)
	}

	if snapshotStatus.Status == "running" {
		fmt.Println("Snapshot is still running...", snapshotID)
		FetchSnapshotStatus(snapshotID)
	}

	return SnapshotStatus{Status: "complete", SnapShotID: snapshotID}, nil

}
