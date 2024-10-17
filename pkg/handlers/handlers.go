package handlers

import (
	"Youtube-Learning-Mode-Ai-Service/pkg/services"
	"encoding/json"
	"fmt"
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

// InitializeAssistantSession: Initialize an assistant with YouTube video metadata and return the assistant ID
func InitializeAssistantSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var initReq services.InitializeRequest
	if err := json.NewDecoder(r.Body).Decode(&initReq); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Create an assistant with metadata
	assistantID, err := services.CreateAssistantWithMetadata(initReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create assistant: %v", err), http.StatusInternalServerError)
		return
	}

	resp := InitializeResponse{AssistantID: assistantID}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
	log.Printf("Assistant initialized with ID: %s for video '%s'", assistantID, initReq.VideoID)
}

// AskAssistantQuestion: Interact with the assistant using the question provided
func AskAssistantQuestion(w http.ResponseWriter, r *http.Request) {
	var req AskAssistantQuestionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	log.Printf("Received question for assistant '%s': %s", req.AssistantID, req.Question)

	// Ask the assistant a question
	answer, err := services.AskAssistantQuestion(req.AssistantID, req.Question)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the assistant's response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AskAssistantResponse{Answer: answer})
}
