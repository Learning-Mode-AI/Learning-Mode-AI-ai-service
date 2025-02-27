package services

import (
	"Learning-Mode-AI-Ai-Service/pkg/config"
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

var (
	RedisClient *redis.Client
	Ctx         = context.Background() // Exported context to be used across the package
)

// Initialize Redis connection
func InitRedis() {
	var tlsConfig *tls.Config
	if config.TLSEnabled {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	} else {
		tlsConfig = nil
	}
	RedisClient = redis.NewClient(&redis.Options{
		Addr:      config.RedisHost, // Redis address
		TLSConfig: tlsConfig,
	})

	err := RedisClient.Ping(Ctx).Err()
	if err != nil {
		panic(err)
	}
}

// GetTranscriptFromRedis retrieves the transcript for a given video ID from Redis
func GetTranscriptFromRedis(videoID string) (string, error) {
	key := videoID
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

func StoreSummaryInRedis(videoID string, summary string) error {
	return RedisClient.Set(Ctx, "summary:"+videoID, summary, 24*time.Hour).Err()
}

func GetSummaryFromRedis(videoID string) (string, error) {
	summary, err := RedisClient.Get(Ctx, "summary:"+videoID).Result()
	if err == redis.Nil {
		return "", nil
	}
	return summary, err
}
