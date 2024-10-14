package handlers

import (
	"Youtube-Learning-Mode-Ai-Service/pkg/services"
	"encoding/json"
	"fmt"
	"net/http"
)

// InitRequest represents the payload for initializing an assistant session
type InitRequest struct {
	VideoID    string   `json:"video_id"`
	Title      string   `json:"title"`
	Channel    string   `json:"channel"`
	Transcript []string `json:"transcript"`
}

// QuestionRequest represents the payload for asking a question to the assistant
type QuestionRequest struct {
	VideoID      string `json:"video_id"`
	UserQuestion string `json:"user_question"`
}

// InitializeAssistantSession initializes a session and stores the assistant and thread IDs
func InitializeAssistantSession(w http.ResponseWriter, r *http.Request) {
	var req InitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	threadID, err := services.CreateAssistantSession(req.VideoID, req.Title, req.Channel, req.Transcript)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to initialize session: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"threadID": threadID})
}

// AskAssistantQuestion handles asking the assistant a question
func AskAssistantQuestion(w http.ResponseWriter, r *http.Request) {
	var req QuestionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	response, err := services.StreamAssistantResponse(req.VideoID, req.UserQuestion)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get response: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"response": response})
}
