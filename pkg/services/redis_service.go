package services

import (
	"Learning-Mode-AI-Ai-Service/pkg/config"
	"context"
	"crypto/tls"
	"fmt"
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
		config.Log.WithFields(map[string]interface{}{
			"error": err.Error(),
			"host":  config.RedisHost,
		}).Fatal("Failed to connect to Redis")
		panic(err)
	}

	config.Log.Info("Successfully connected to Redis")
}

// GetTranscriptFromRedis retrieves the transcript for a given video ID from Redis
func GetTranscriptFromRedis(videoID string) (string, error) {
	key := videoID
	config.Log.WithFields(map[string]interface{}{
		"video_id": videoID,
	}).Debug("Querying Redis for transcript")

	val, err := RedisClient.Get(Ctx, key).Result()
	if err == redis.Nil {
		config.Log.WithFields(map[string]interface{}{
			"video_id": videoID,
		}).Info("Transcript not found in Redis")
		return "", nil
	} else if err != nil {
		config.Log.WithFields(map[string]interface{}{
			"video_id": videoID,
			"error":    err.Error(),
		}).Error("Error retrieving transcript from Redis")
		return "", fmt.Errorf("error retrieving from Redis: %v", err)
	}

	config.Log.WithFields(map[string]interface{}{
		"video_id": videoID,
	}).Debug("Transcript successfully retrieved from Redis")
	return val, nil
}

func StoreSummaryInRedis(videoID string, summary string) error {
	config.Log.WithFields(map[string]interface{}{
		"video_id": videoID,
	}).Debug("Storing summary in Redis")

	err := RedisClient.Set(Ctx, "summary:"+videoID, summary, 24*time.Hour).Err()
	if err != nil {
		config.Log.WithFields(map[string]interface{}{
			"video_id": videoID,
			"error":    err.Error(),
		}).Error("Failed to store summary in Redis")
		return err
	}

	config.Log.WithFields(map[string]interface{}{
		"video_id": videoID,
	}).Debug("Summary successfully stored in Redis")
	return nil
}

func GetSummaryFromRedis(videoID string) (string, error) {
	config.Log.WithFields(map[string]interface{}{
		"video_id": videoID,
	}).Debug("Retrieving summary from Redis")

	summary, err := RedisClient.Get(Ctx, "summary:"+videoID).Result()
	if err == redis.Nil {
		config.Log.WithFields(map[string]interface{}{
			"video_id": videoID,
		}).Info("Summary not found in Redis")
		return "", nil
	} else if err != nil {
		config.Log.WithFields(map[string]interface{}{
			"video_id": videoID,
			"error":    err.Error(),
		}).Error("Error retrieving summary from Redis")
		return "", err
	}

	config.Log.WithFields(map[string]interface{}{
		"video_id": videoID,
	}).Debug("Summary successfully retrieved from Redis")
	return summary, nil
}

func GetAssistantIDFromRedis(userID, videoID string) (string, error) {
	ctx := context.Background()
	redisKey := fmt.Sprintf("assistant:%s:%s", userID, videoID)

	config.Log.WithFields(map[string]interface{}{
		"user_id":  userID,
		"video_id": videoID,
	}).Debug("Looking up assistant ID in Redis")

	assistantID, err := RedisClient.Get(ctx, redisKey).Result()
	if err == redis.Nil {
		config.Log.WithFields(map[string]interface{}{
			"user_id":  userID,
			"video_id": videoID,
		}).Warn("AssistantID not found in Redis")
		return "", fmt.Errorf("AssistantID not found in Redis for UserID: %s and VideoID: %s", userID, videoID)
	} else if err != nil {
		config.Log.WithFields(map[string]interface{}{
			"user_id":  userID,
			"video_id": videoID,
			"error":    err.Error(),
		}).Error("Redis error when looking up assistant ID")
		return "", fmt.Errorf("Redis error: %v", err)
	}

	config.Log.WithFields(map[string]interface{}{
		"user_id":      userID,
		"video_id":     videoID,
		"assistant_id": assistantID,
	}).Debug("Successfully retrieved assistant ID from Redis")
	return assistantID, nil
}
