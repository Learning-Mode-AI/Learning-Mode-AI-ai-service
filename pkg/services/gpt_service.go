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
	// Create the initial system message with video context
	initialMessage := fmt.Sprintf(
		"You are helping a user based on the following video:\n\nTitle: %s\nChannel: %s\nTranscript:\n%s",
		title, channel, strings.Join(transcript, "\n"),
	)

	// Store this initial message in Redis
	err := redisClient.HSet(ctx, videoID, "conversation", initialMessage).Err()
	if err != nil {
		return fmt.Errorf("failed to store GPT session base: %v", err)
	}

	log.Printf("GPT session initialized for video ID: %s", videoID)
	return nil
}

// FetchGPTResponse generates a response from GPT-4 based on a user's question
func FetchGPTResponse(videoID, userQuestion string) (string, error) {
	// Retrieve the existing conversation history from Redis
	conversation, err := redisClient.HGet(ctx, videoID, "conversation").Result()
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
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    openai.GPT4,
			Messages: messages,
		},
	)
	if err != nil {
		return "", fmt.Errorf("GPT-4 request failed: %v", err)
	}

	// Append the AI response to the conversation history and store in Redis
	aiResponse := resp.Choices[0].Message.Content
	updatedConversation := conversation + fmt.Sprintf("\nUser: %s\nAI: %s", userQuestion, aiResponse)

	err = redisClient.HSet(ctx, videoID, "conversation", updatedConversation).Err()
	if err != nil {
		return "", fmt.Errorf("failed to store updated conversation in Redis: %v", err)
	}

	return aiResponse, nil
}
