package handlers

import (
	"encoding/json"
	"net/http"
	"Learning-Mode-AI-Ai-Service/pkg/services"
	"log"
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

    // Debug log to confirm correct video ID
    log.Printf("Request received with video ID: %s", req.VideoID)

    // Retrieve the transcript from Redis
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

    // Generate the summary using the transcript
    summary, err := services.GenerateSummary(transcript)
    if err != nil {
        log.Printf("Error generating summary: %v", err)
        http.Error(w, "Failed to generate summary", http.StatusInternalServerError)
        return
    }

    // Respond with the generated summary
    resp := SummaryResponse{Summary: summary}
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
}

