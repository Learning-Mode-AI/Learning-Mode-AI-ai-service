package main

import (
	"context"
	"fmt"
	"log"
	"os"

	openai "github.com/sashabaranov/go-openai"
)

var openaiClient *openai.Client

func initOpenAIClient() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OpenAI API key is missing")
	}

	openaiClient = openai.NewClient(apiKey)
}

func main() {
	// Initialize OpenAI client
	initOpenAIClient()

	// Example usage of the OpenAI client
	req := openai.CompletionRequest{
		Model:  openai.GPT3TextDavinci003,
		Prompt: "What is Go programming language?",
	}

	resp, err := openaiClient.CreateCompletion(context.Background(), req)
	if err != nil {
		log.Fatalf("OpenAI API call failed: %v", err)
	}

	fmt.Println(resp.Choices[0].Text)
}
