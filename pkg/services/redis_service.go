package services

import (
	"context"
	"fmt"
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
		Addr:     "localhost:6379", // Redis address
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

// GetThreadID retrieves the thread ID for a given video ID from Redis
func GetThreadID(videoID string) (string, error) {
	threadID, err := RedisClient.Get(Ctx, fmt.Sprintf("%s:threadID", videoID)).Result()
	if err == redis.Nil {
		log.Printf("No thread ID found for video ID: %s", videoID)
		return "", fmt.Errorf("no thread ID found for video ID: %s", videoID)
	} else if err != nil {
		log.Printf("Error retrieving thread ID: %v", err)
		return "", fmt.Errorf("error retrieving thread ID: %v", err)
	}

	log.Printf("Fetched thread ID: %s", threadID)
	return threadID, nil
}

func GetAssistantID(videoID string) (string, error) {
	assistantID, err := RedisClient.Get(Ctx, fmt.Sprintf("%s:assistantID", videoID)).Result()
	if err == redis.Nil {
		log.Printf("No assistant ID found for video ID: %s", videoID)
		return "", fmt.Errorf("no assistant ID found for video ID: %s", videoID)
	} else if err != nil {
		log.Printf("Error retrieving assistant ID: %v", err)
		return "", fmt.Errorf("error retrieving assistant ID: %v", err)
	}
	return assistantID, nil
}
