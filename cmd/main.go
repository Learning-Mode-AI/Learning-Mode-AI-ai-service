package main

import (
	"log"
	"net/http"

	"Learning-Mode-AI-Ai-Service/pkg/config"
	"Learning-Mode-AI-Ai-Service/pkg/handlers"
	"Learning-Mode-AI-Ai-Service/pkg/services"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	config.InitConfig()
	services.InitRedis()
}

func main() {
	// Set up router
	r := mux.NewRouter()

	// Define routes
	r.HandleFunc("/ai/init-session", handlers.InitializeAssistantSession).Methods("POST")
	r.HandleFunc("/ai/ask-question", handlers.AskAssistantQuestion).Methods("POST")

	// Start the server
	log.Println("AI Service running on :8082")
	log.Fatal(http.ListenAndServe(":8082", r))
}
