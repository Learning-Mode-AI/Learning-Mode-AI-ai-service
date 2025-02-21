package config

import (
	"fmt"
	"os"
)

var (
	RedisHost string
)

func InitConfig() {
	env := os.Getenv("ENVIRONMENT")
	if env == "local" {
		RedisHost = "localhost"
		fmt.Println("Running in local mode")
	} else {
		redisEnvHost = os.Getenv("REDIS_HOST")
		if redisEnvHost != "" {
			RedisHost = redisEnvHost
		} else {
			RedisHost = "redis:6379"
		}
		fmt.Println("Running in Docker mode")
	}
}
