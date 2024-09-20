package handlers

import (
	"Youtube-Learning-Mode-Ai-Service/pkg/services"
	"encoding/json"
	"net/http"
)

// Request to initialize a GPT session with video context
type InitRequest struct {
	VideoID    string   `json:"video_id"`
	Title      string   `json:"title"`
	Channel    string   `json:"channel"`
	Transcript []string `json:"transcript"`
}

// Request for asking GPT questions
type QuestionRequest struct {
	VideoID      string `json:"video_id"`
	UserQuestion string `json:"user_question"`
}

// Initialize GPT session with video context
func InitializeGPTSession(w http.ResponseWriter, r *http.Request) {
	var initReq InitRequest
	if err := json.NewDecoder(r.Body).Decode(&initReq); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Call service to initialize GPT session with transcript
	err := services.CreateGPTSession(initReq.VideoID, initReq.Title, initReq.Channel, initReq.Transcript)
	if err != nil {
		http.Error(w, "Failed to initialize GPT session", http.StatusInternalServerError)
		return
	}

	// Respond to client
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "GPT session initialized"})
}

// Handle user questions
func AskGPTQuestion(w http.ResponseWriter, r *http.Request) {
	var questionReq QuestionRequest
	if err := json.NewDecoder(r.Body).Decode(&questionReq); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Get GPT response
	aiResponse, err := services.FetchGPTResponse(questionReq.VideoID, questionReq.UserQuestion)
	if err != nil {
		http.Error(w, "Failed to get AI response", http.StatusInternalServerError)
		return
	}

	// Respond with AI answer
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"response": aiResponse})
}
