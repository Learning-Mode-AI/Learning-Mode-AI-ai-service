package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"Youtube-Learning-Mode-Ai-Service/pkg/handlers"
	"Youtube-Learning-Mode-Ai-Service/pkg/services"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	openai "github.com/sashabaranov/go-openai"
)

var openaiClient *openai.Client

func initOpenAIClient() {
	envPath := filepath.Join("..", ".env")
	err := godotenv.Load(envPath)
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OpenAI API key is missing")
	}

	openaiClient = openai.NewClient(apiKey)
}

func main() {
	// Initialize OpenAI client and Redis client
	initOpenAIClient()
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
