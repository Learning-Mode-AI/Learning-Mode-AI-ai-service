package handlers

import (
	"Learning-Mode-AI-Ai-Service/pkg/services"
	"encoding/json"
	"log"
	"net/http"
)

type InitializeResponse struct {
	AssistantID string `json:"assistant_id"`
}

type AskAssistantQuestionRequest struct {
	AssistantID string `json:"assistant_id"`
	Question    string `json:"question"`
}

type AskAssistantResponse struct {
	Answer string `json:"answer"`
	Error  string `json:"error,omitempty"`
}

// InitializeAssistantSession: Create a new assistant based on YouTube video metadata and return the assistant ID.
func InitializeAssistantSession(w http.ResponseWriter, r *http.Request) {
	// Decode the incoming request
	var initReq services.InitializeRequest
	if err := json.NewDecoder(r.Body).Decode(&initReq); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Create an assistant with metadata
	assistantID, err := services.CreateAssistantWithMetadata(initReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"message":      "Assistant session initialized successfully.",
		"assistant_id": assistantID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	log.Printf("Assistant session initialized with ID: %s for video '%s'", assistantID, initReq.VideoID)
}

// Handler for asking a question to the assistant
func AskAssistantQuestion(w http.ResponseWriter, r *http.Request) {
	var req struct {
		VideoID     string `json:"video_id"`
		AssistantID string `json:"assistant_id"`
		Question    string `json:"question"`
		Timestamp   int    `json:"timestamp"`
	}

	// Parse the request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	log.Printf("Received question for assistant '%s': %s at timestamp: %d", req.AssistantID, req.Question, req.Timestamp)

	// Pass the timestamp to the service
	response, err := services.AskAssistantQuestion(req.VideoID, req.AssistantID, req.Question, req.Timestamp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the assistant's response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"answer": response})
}
