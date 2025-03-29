package handlers

import (
	"Learning-Mode-AI-Ai-Service/pkg/config"
	"Learning-Mode-AI-Ai-Service/pkg/services"
	"encoding/json"
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
	config.Log.WithFields(map[string]interface{}{
		"video_id":     initReq.VideoID,
		"assistant_id": assistantID,
	}).Info("Assistant session initialized successfully")
}

// Handler for asking a question to the assistant
func AskAssistantQuestion(w http.ResponseWriter, r *http.Request) {
	var req struct {
		VideoID   string `json:"video_id"`
		UserID    string `json:"userId"`
		Question  string `json:"question"`
		Timestamp int    `json:"timestamp"`
	}

	// Parse the request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	config.Log.WithFields(map[string]interface{}{
		"user_id":  req.UserID,
		"video_id": req.VideoID,
	}).Info("Looking up AssistantID for user and video")

	assistantID, err := services.GetAssistantIDFromRedis(req.UserID, req.VideoID)
	if err != nil {
		http.Error(w, "Assistant session not found for this user and video", http.StatusBadRequest)
		return
	}

	config.Log.WithFields(map[string]interface{}{
		"user_id":      req.UserID,
		"video_id":     req.VideoID,
		"assistant_id": assistantID,
	}).Info("Found AssistantID")

	// Pass the timestamp to the service
	response, err := services.AskAssistantQuestion(req.VideoID, assistantID, req.Question, req.Timestamp)
	if err != nil {
		config.Log.WithFields(map[string]interface{}{
			"user_id":      req.UserID,
			"video_id":     req.VideoID,
			"assistant_id": assistantID,
			"error":        err.Error(),
		}).Error("Failed to get answer from assistant")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the assistant's response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"answer": response})

	config.Log.WithFields(map[string]interface{}{
		"user_id":      req.UserID,
		"video_id":     req.VideoID,
		"assistant_id": assistantID,
	}).Info("Successfully received answer from assistant")
}
