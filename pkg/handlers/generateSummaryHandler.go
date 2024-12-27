package handlers

import (
	"encoding/json"
	"net/http"
	"Learning-Mode-AI-Ai-Service/pkg/services"
)

type SummaryRequest struct {
	Transcript string `json:"transcript"`
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

	// Call the GPT service to generate a summary
	summary, err := services.GenerateSummary(req.Transcript)
	if err != nil {
		http.Error(w, "Failed to generate summary", http.StatusInternalServerError)
		return
	}

	resp := SummaryResponse{Summary: summary}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
