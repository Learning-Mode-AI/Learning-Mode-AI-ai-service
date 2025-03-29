package handlers

import (
	"Learning-Mode-AI-Ai-Service/pkg/config"
	"Learning-Mode-AI-Ai-Service/pkg/services"
	"encoding/json"
	"net/http"
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

	config.Log.WithFields(map[string]interface{}{
		"video_id": req.VideoID,
	}).Info("Quiz generation request received")

	transcript, err := services.GetTranscriptFromRedis(req.VideoID)
	if err != nil {
		config.Log.WithFields(map[string]interface{}{
			"video_id": req.VideoID,
			"error":    err.Error(),
		}).Error("Failed to retrieve transcript from Redis")
		http.Error(w, "Failed to retrieve transcript", http.StatusInternalServerError)
		return
	}

	if transcript == "" {
		config.Log.WithFields(map[string]interface{}{
			"video_id": req.VideoID,
		}).Warn("Transcript not found in Redis")
		http.Error(w, "Transcript not found", http.StatusNotFound)
		return
	}

	quiz, err := services.GenerateQuiz(transcript)
	if err != nil {
		config.Log.WithFields(map[string]interface{}{
			"video_id": req.VideoID,
			"error":    err.Error(),
		}).Error("Failed to generate quiz")
		http.Error(w, "Failed to generate quiz", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(quiz); err != nil {
		config.Log.WithFields(map[string]interface{}{
			"video_id": req.VideoID,
			"error":    err.Error(),
		}).Error("Failed to encode quiz response")
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}

	config.Log.WithFields(map[string]interface{}{
		"video_id": req.VideoID,
	}).Info("Quiz successfully generated")
}
