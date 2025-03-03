package config

import (
	"fmt"
	"os"
	"strconv"
)

var (
	RedisHost  string
	TLSEnabled bool
)

func InitConfig() {
	env := os.Getenv("ENVIRONMENT")
	tlsEnv := os.Getenv("TLS_ENABLED")
	tlsEnabled, err := strconv.ParseBool(tlsEnv)
	if err != nil {
		tlsEnabled = false
	}
	TLSEnabled = tlsEnabled

	if env == "local" {
		RedisHost = "localhost:6379"
		TLSEnabled = false
		fmt.Println("Running in local mode")
	} else {
		redisEnvHost := os.Getenv("REDIS_HOST")
		if redisEnvHost != "" {
			RedisHost = redisEnvHost
		} else {
			RedisHost = "redis:6379"
		}
		fmt.Println("Running in Docker mode")
	}
}
