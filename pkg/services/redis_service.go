package services

import (
	"Learning-Mode-AI-Ai-Service/pkg/config"
	"context"
	"log"
	"fmt"
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

// GetTranscriptFromRedis retrieves the transcript for a given video ID from Redis
func GetTranscriptFromRedis(videoID string) (string, error) {
    key := "transcript:" + videoID
    log.Printf("Querying Redis with key: %s", key)
    val, err := RedisClient.Get(Ctx, key).Result()
    if err == redis.Nil {
        log.Printf("Transcript not found for key: %s", key)
        return "", nil
    } else if err != nil {
        return "", fmt.Errorf("error retrieving from Redis: %v", err)
    }
    return val, nil
}

// StoreTranscriptInRedis stores a transcript in Redis for a given video ID
func StoreTranscriptInRedis(videoID, transcript string) error {
	key := "transcript:" + videoID
	err := RedisClient.Set(Ctx, key, transcript, 0).Err() // No expiration
	if err != nil {
		log.Printf("Error storing transcript in Redis for video ID: %s, %v", videoID, err)
		return err
	}
	log.Printf("Transcript stored in Redis for video ID: %s", videoID)
	return nil
}