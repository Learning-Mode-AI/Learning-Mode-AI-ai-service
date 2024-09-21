package main

import (
	"log"
	"net/http"

	"Youtube-Learning-Mode-Ai-Service/pkg/handlers"
	"Youtube-Learning-Mode-Ai-Service/pkg/services"

	"github.com/gorilla/mux"
)

func main() {
	// Initialize OpenAI client and Redis client
	services.InitOpenAIClient()
	services.InitRedis()

	// Set up router
	r := mux.NewRouter()

	// Define routes
	r.HandleFunc("/ai/init-session", handlers.InitializeGPTSession).Methods("POST")
	r.HandleFunc("/ai/ask-question", handlers.AskGPTQuestion).Methods("POST")

	// Start the server
	log.Println("AI Service running on :8082")
	log.Fatal(http.ListenAndServe(":8082", r))
}
