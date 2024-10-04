package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
)

type OpenAIRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	Logprobs     *string `json:"logprobs"`
	FinishReason string  `json:"finish_reason"`
}

type CompletionTokensDetails struct {
	ReasoningTokens int `json:"reasoning_tokens"`
}

type Usage struct {
	PromptTokens            int                     `json:"prompt_tokens"`
	CompletionTokens        int                     `json:"completion_tokens"`
	TotalTokens             int                     `json:"total_tokens"`
	CompletionTokensDetails CompletionTokensDetails `json:"completion_tokens_details"`
}

type ChatCompletionResponse struct {
	ID                string   `json:"id"`
	Object            string   `json:"object"`
	Created           int64    `json:"created"`
	Model             string   `json:"model"`
	SystemFingerprint string   `json:"system_fingerprint"`
	Choices           []Choice `json:"choices"`
	Usage             Usage    `json:"usage"`
}

type AI interface {
	Summarize(text string) (string, error)
}

type Summary string

type ScrapedData struct {
	Summary
}

func (ai *ScrapedData) Summarize(text string) (Summary, error) {

	client := &http.Client{}
	// Create the request body
	openAIRequest := OpenAIRequest{
		Model: "gpt-4o",
		Messages: []Message{
			Message{
				Role:    "user",
				Content: "I want you to summarize the what information is on following webpage in text format.",
			},
			Message{
				Role:    "user",
				Content: text,
			},
		},
	}

	// Parse openAIRequest to JSON
	jsonData, err := json.Marshal(openAIRequest)
	if err != nil {
		return ai.Summary, err
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return ai.Summary, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("OPENAI_API_KEY"))

	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error sending request to OpenAI:", err)
		return ai.Summary, err
	}

	defer resp.Body.Close()

	var completionResponse ChatCompletionResponse
	err = json.NewDecoder(resp.Body).Decode(&completionResponse)
	if err != nil {
		log.Println("Error decoding response from OpenAI:", err)
		return ai.Summary, err
	}

	ai.Summary = Summary(completionResponse.Choices[0].Message.Content)
	return ai.Summary, nil
}
