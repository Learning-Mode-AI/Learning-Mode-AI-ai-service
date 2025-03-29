package config

import (
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
)

var (
	RedisHost  string
	TLSEnabled bool
	Log        *logrus.Logger
)

func InitConfig() {
	// Initialize logrus
	Log = logrus.New()
	Log.SetFormatter(&logrus.JSONFormatter{})
	Log.SetOutput(os.Stdout)

	// Set log level based on environment
	logLevel := os.Getenv("LOG_LEVEL")
	switch logLevel {
	case "debug":
		Log.SetLevel(logrus.DebugLevel)
	case "info":
		Log.SetLevel(logrus.InfoLevel)
	case "warn":
		Log.SetLevel(logrus.WarnLevel)
	case "error":
		Log.SetLevel(logrus.ErrorLevel)
	default:
		Log.SetLevel(logrus.InfoLevel)
	}

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
		Log.Info("Running in local mode")
	} else {
		redisEnvHost := os.Getenv("REDIS_HOST")
		if redisEnvHost != "" {
			RedisHost = redisEnvHost
		} else {
			RedisHost = "redis:6379"
		}
		Log.Info("Running in Docker mode")
	}
}
