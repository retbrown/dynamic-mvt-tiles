package config

import (
	"os"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

type Config struct {
	ConnStr string
}

// Gather the config struct
func Gather(logger *zap.Logger) *Config {
	if err := godotenv.Load(); err != nil {
		logger.Debug("No .env file found")
	}
	return &Config{
		ConnStr: getEnv(logger, "CONN_STRING", false),
	}
}

// Simple helper function to read an environment variable
func getEnv(logger *zap.Logger, key string, optional bool) string {
	value, exists := os.LookupEnv(key)

	if !exists && optional {
		logger.Info("Variable not found", zap.String("variable", key))
	} else if !exists {
		logger.Fatal("Unable to find variable", zap.String("variable", key))
	}
	return value
}
