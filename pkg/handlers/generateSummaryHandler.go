package handlers

import (
	"encoding/json"
	"net/http"
	"Learning-Mode-AI-Ai-Service/pkg/services"
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

	// Check Redis for existing summary
	summary, err := services.GetSummaryFromRedis(req.VideoID)
	if err != nil {
		http.Error(w, "Error checking cache", http.StatusInternalServerError)
		return
	}

	if summary != "" {
		resp := SummaryResponse{Summary: summary}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	// If not cached, generate a new summary
	transcript, err := services.GetTranscriptFromRedis(req.VideoID)
	if transcript == "" {
		http.Error(w, "Transcript not found", http.StatusNotFound)
		return
	}

	summary, err = services.GenerateSummary(transcript)
	if err != nil {
		http.Error(w, "Failed to generate summary", http.StatusInternalServerError)
		return
	}

	// Cache the new summary in Redis
	services.StoreSummaryInRedis(req.VideoID, summary)

	resp := SummaryResponse{Summary: summary}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}