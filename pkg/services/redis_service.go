package services

import (
	"Learning-Mode-AI-Ai-Service/pkg/config"
	"context"
	"log"

	"github.com/go-redis/redis/v8"
)

var (
	RedisClient *redis.Client
	Ctx         = context.Background() // Exported context to be used across the package
)

// Initialize Redis connection
func InitRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     config.RedisHost, // Redis address
		Password: "",               // No password set
		DB:       0,                // Use default DB
	})

	// Test the connection
	_, err := RedisClient.Ping(Ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Connected to Redis")
}
