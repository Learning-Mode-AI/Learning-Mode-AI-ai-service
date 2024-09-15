package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	openai "github.com/sashabaranov/go-openai"
)

type AIChat struct {
	Client   *openai.Client
	Messages []openai.ChatCompletionMessage
}

type AIRequest struct {
	VideoInfo    VideoInfo `json:"video_info"`
	UserQuestion string    `json:"user_question"`
}

type VideoInfo struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Channel     string `json:"channel"`
	Transcript  string `json:"transcript"`
}

type AIResponse struct {
	Response string `json:"response"`
}

var openaiClient *openai.Client

// Handler to process AI question requests
func handleAIRequest(w http.ResponseWriter, r *http.Request) {
	var aiRequest AIRequest
	err := json.NewDecoder(r.Body).Decode(&aiRequest)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Initialize the AIChat instance
	aiChat, err := CreateChatGPTInstance(&aiRequest.VideoInfo)
	if err != nil {
		http.Error(w, "Failed to initialize AI session", http.StatusInternalServerError)
		return
	}

	// Generate response based on user question
	aiResponse, err := aiChat.FetchGPTResponse(aiRequest.UserQuestion)
	if err != nil {
		http.Error(w, "Failed to generate AI response", http.StatusInternalServerError)
		return
	}

	// Send the AI response back
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AIResponse{Response: aiResponse})
}

// CreateChatGPTInstance initializes a ChatGPT instance with the video context.
func CreateChatGPTInstance(videoInfo *VideoInfo) (*AIChat, error) {
	// Initialize the ChatGPT session with the video context
	initialMessage := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: fmt.Sprintf("You are helping a user based on the following video:\n\nTitle: %s\nDescription: %s\nChannel: %s\nTranscript:\n%s", videoInfo.Title, videoInfo.Description, videoInfo.Channel, videoInfo.Transcript),
	}

	return &AIChat{
		Client:   openaiClient,
		Messages: []openai.ChatCompletionMessage{initialMessage},
	}, nil
}

// FetchGPTResponse generates a response from GPT-4 based on the provided user question.
func (ai *AIChat) FetchGPTResponse(userQuestion string) (string, error) {
	// Append the user question to the conversation history
	ai.Messages = append(ai.Messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: userQuestion,
	})

	// Call the OpenAI API with the entire conversation history
	resp, err := ai.Client.CreateChatCompletion(
		context.Background(), // Corrected here
		openai.ChatCompletionRequest{
			Model:    openai.GPT4,
			Messages: ai.Messages,
		},
	)

	if err != nil {
		return "", fmt.Errorf("ChatCompletion error: %v", err)
	}

	// Append the AI response to the conversation history
	ai.Messages = append(ai.Messages, resp.Choices[0].Message)

	// Return the generated response
	return resp.Choices[0].Message.Content, nil
}

func main() {
	// Initialize OpenAI client
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OpenAI API key is missing")
	}

	openaiClient = openai.NewClient(apiKey)

	http.HandleFunc("/ai", handleAIRequest)

	log.Println("AI Service running on :8082")
	log.Fatal(http.ListenAndServe(":8082", nil))
}
