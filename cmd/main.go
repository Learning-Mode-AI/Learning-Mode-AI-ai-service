package main

import (
	"log"
	"net/http"

	"Learning-Mode-AI-Ai-Service/pkg/handlers"
	"Learning-Mode-AI-Ai-Service/pkg/services"

	"github.com/gorilla/mux"
)

func main() {
	// Initialize OpenAI client and Redis connection
	services.InitOpenAIClient()
	services.InitRedis()

	// Set up router
	r := mux.NewRouter()

	// Define routes
	r.HandleFunc("/ai/init-session", handlers.InitializeAssistantSession).Methods("POST")
	r.HandleFunc("/ai/ask-question", handlers.AskAssistantQuestion).Methods("POST")

	// Start the server
	log.Println("AI Service running on :8082")
	log.Fatal(http.ListenAndServe(":8082", r))
}
