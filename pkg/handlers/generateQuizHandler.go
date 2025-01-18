package handlers

import (
	"encoding/json"
	"net/http"
	"Learning-Mode-AI-Ai-Service/pkg/services"
	"log"
)

type QuizRequest struct {
	VideoID string `json:"video_id"` // Accept only the video_id
}

type QuizResponse struct {
	Quiz map[string]interface{} `json:"quiz"`
}

func GenerateQuizHandler(w http.ResponseWriter, r *http.Request) {
	var req QuizRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	log.Printf("Request received with video ID: %s", req.VideoID)

	transcript, err := services.GetTranscriptFromRedis(req.VideoID)
	if err != nil {
		log.Printf("Error retrieving transcript from Redis: %v", err)
		http.Error(w, "Failed to retrieve transcript", http.StatusInternalServerError)
		return
	}

	if transcript == "" {
		log.Printf("Transcript not found in Redis for video ID: %s", req.VideoID)
		http.Error(w, "Transcript not found", http.StatusNotFound)
		return
	}

	quiz, err := services.GenerateQuiz(transcript)
	if err != nil {
		log.Printf("Error generating quiz: %v", err)
		http.Error(w, "Failed to generate quiz", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(quiz); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
