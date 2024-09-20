package services

import (
	"context"
	"fmt"
	"log"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

var openaiClient *openai.Client

// Create a GPT session with video info
func CreateGPTSession(videoID, title, channel string, transcript []string) error {
	ctx := context.Background()

	// Create the initial system message with video context
	initialMessage := fmt.Sprintf(
		"You are helping a user based on the following video:\n\nTitle: %s\nChannel: %s\nTranscript:\n%s",
		title, channel, strings.Join(transcript, "\n"),
	)

	gptKey := fmt.Sprintf("%s:conversation", videoID)
	log.Printf("Attempting to store GPT conversation using key: %s", gptKey)

	// Use SETNX to create the session only if it doesn't exist
	success, err := redisClient.SetNX(ctx, gptKey, initialMessage, 0).Result()
	if err != nil {
		return fmt.Errorf("Redis SETNX failed: %v", err)
	}

	if success {
		log.Printf("GPT session initialized for video ID: %s", videoID)
	} else {
		log.Printf("GPT session already exists for video ID: %s", videoID)
	}

	return nil
}

// FetchGPTResponse generates a response from GPT-4 based on a user's question
// FetchGPTResponse generates a response from GPT-4 based on a user's question
func FetchGPTResponse(videoID, userQuestion string) (string, error) {
	ctx := context.Background()

	// Ensure OpenAI client is initialized
	if openaiClient == nil {
		return "", fmt.Errorf("OpenAI client is not initialized")
	}

	// Retrieve the existing conversation history from Redis
	conversation, err := redisClient.Get(ctx, fmt.Sprintf("%s:conversation", videoID)).Result()
	if err != nil {
		return "", fmt.Errorf("failed to retrieve conversation from Redis: %v", err)
	}

	// Append the user question to the conversation
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: conversation,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: userQuestion,
		},
	}

	// Call GPT-4 with the conversation and the new question
	resp, err := openaiClient.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:    openai.GPT4,
			Messages: messages,
		},
	)
	if err != nil {
		return "", fmt.Errorf("GPT-4 request failed: %v", err)
	}

	// Append the AI response to the conversation history and store it back in Redis
	aiResponse := resp.Choices[0].Message.Content
	updatedConversation := fmt.Sprintf("%s\nUser: %s\nAI: %s", conversation, userQuestion, aiResponse)

	err = redisClient.Set(ctx, fmt.Sprintf("%s:conversation", videoID), updatedConversation, 0).Err()
	if err != nil {
		return "", fmt.Errorf("failed to store updated conversation in Redis: %v", err)
	}

	return aiResponse, nil
}
