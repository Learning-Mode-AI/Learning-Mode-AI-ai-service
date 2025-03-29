package handlers

import (
	"Learning-Mode-AI-Ai-Service/pkg/config"
	"Learning-Mode-AI-Ai-Service/pkg/services"
	"encoding/json"
	"net/http"
)

type SummaryRequest struct {
	VideoID string `json:"video_id"` // Accept only the video_id
}

type SummaryResponse struct {
	Summary string `json:"summary"`
}

func GenerateSummaryHandler(w http.ResponseWriter, r *http.Request) {
	var req SummaryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	config.Log.WithFields(map[string]interface{}{
		"video_id": req.VideoID,
	}).Info("Summary generation request received")

	// Check Redis for existing summary
	summary, err := services.GetSummaryFromRedis(req.VideoID)
	if err != nil {
		config.Log.WithFields(map[string]interface{}{
			"video_id": req.VideoID,
			"error":    err.Error(),
		}).Error("Error checking cache for summary")
		http.Error(w, "Error checking cache", http.StatusInternalServerError)
		return
	}

	if summary != "" {
		config.Log.WithFields(map[string]interface{}{
			"video_id": req.VideoID,
		}).Info("Using cached summary")
		resp := SummaryResponse{Summary: summary}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	// If not cached, generate a new summary
	transcript, err := services.GetTranscriptFromRedis(req.VideoID)
	if transcript == "" {
		config.Log.WithFields(map[string]interface{}{
			"video_id": req.VideoID,
		}).Warn("Transcript not found")
		http.Error(w, "Transcript not found", http.StatusNotFound)
		return
	}

	config.Log.WithFields(map[string]interface{}{
		"video_id": req.VideoID,
	}).Info("Generating new summary")

	summary, err = services.GenerateSummary(transcript)
	if err != nil {
		config.Log.WithFields(map[string]interface{}{
			"video_id": req.VideoID,
			"error":    err.Error(),
		}).Error("Failed to generate summary")
		http.Error(w, "Failed to generate summary", http.StatusInternalServerError)
		return
	}

	// Cache the new summary in Redis
	if err := services.StoreSummaryInRedis(req.VideoID, summary); err != nil {
		config.Log.WithFields(map[string]interface{}{
			"video_id": req.VideoID,
			"error":    err.Error(),
		}).Warn("Failed to cache summary in Redis")
	} else {
		config.Log.WithFields(map[string]interface{}{
			"video_id": req.VideoID,
		}).Info("Summary cached in Redis")
	}

	resp := SummaryResponse{Summary: summary}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

	config.Log.WithFields(map[string]interface{}{
		"video_id": req.VideoID,
	}).Info("Summary successfully generated and returned")
}
