package services

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	openai "github.com/sashabaranov/go-openai"
)

// Initialize the OpenAI client globally
var OpenAIClient *openai.Client

// GetEnv retrieves environment variables with a fallback value.
func GetEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// InitOpenAIClient initializes the OpenAI client
func InitOpenAIClient() {
	envPath := "../.env" // or the correct relative path
	if err := godotenv.Load(envPath); err != nil {
		log.Fatalf("Error loading .env file from %s: %v", envPath, err)
	}

	apiKey := GetEnv("OPENAI_API_KEY", "")
	if apiKey == "" {
		log.Fatal("OpenAI API key is missing")
	}
	log.Println("OpenAI API Key loaded successfully")

	OpenAIClient = openai.NewClient(apiKey)
}

// CreateAssistantSession creates an assistant and stores IDs in Redis
func CreateAssistantSession(videoID, title, channel string, transcript []string) (string, error) {
	ctx := context.Background()

	// Create the assistant
	req := openai.AssistantRequest{
		Model:        "gpt-4",
		Name:         stringPtr("YouTube Learning Mode Assistant"),
		Instructions: stringPtr(fmt.Sprintf("Help users with video: %s (%s)\n%s", title, channel, strings.Join(transcript, "\n"))),
	}

	assistant, err := OpenAIClient.CreateAssistant(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to create assistant: %v", err)
	}

	// Store assistant ID in Redis
	if err := RedisClient.Set(Ctx, fmt.Sprintf("%s:assistantID", videoID), assistant.ID, 0).Err(); err != nil {
		return "", fmt.Errorf("failed to store assistant ID: %v", err)
	}

	// Create a thread and store the thread ID in Redis
	thread, err := OpenAIClient.CreateThread(ctx, openai.ThreadRequest{})
	if err != nil {
		return "", fmt.Errorf("failed to create thread: %v", err)
	}
	if err := RedisClient.Set(Ctx, fmt.Sprintf("%s:threadID", videoID), thread.ID, 0).Err(); err != nil {
		return "", fmt.Errorf("failed to store thread ID: %v", err)
	}

	return thread.ID, nil
}

// StreamAssistantResponse sends a message and streams the assistant's response
func StreamAssistantResponse(videoID, userQuestion string) (string, error) {
	ctx := context.Background()

	threadID, err := GetThreadID(videoID)
	if err != nil {
		return "", err
	}

	// Send the user’s question to the thread
	resp, err := OpenAIClient.CreateMessage(ctx, threadID, openai.MessageRequest{
		Role:    string(openai.ThreadMessageRoleUser),
		Content: userQuestion,
	})
	if err != nil {
		return "", fmt.Errorf("failed to send message: %v", err)
	}

	// Return the assistant's response
	if len(resp.Content) > 0 && resp.Content[0].Text != nil {
		return resp.Content[0].Text.Value, nil
	}

	return "", fmt.Errorf("no response from assistant")
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
