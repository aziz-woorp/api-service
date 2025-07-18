package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort   string
	AppEnv    string
	GinMode   string
	LogLevel  string
	MongoURI  string
}

func LoadConfig() *Config {
	// Load .env if present
	_ = godotenv.Load(".env")

	cfg := &Config{
		AppPort:  getEnv("APP_PORT", "8080"),
		AppEnv:   getEnv("APP_ENV", "development"),
		GinMode:  getEnv("GIN_MODE", "debug"),
		LogLevel: getEnv("LOG_LEVEL", "INFO"),
		MongoURI: getEnv("MONGO_URI", "mongodb://localhost:27017/api_service_dev"),
	}

	return cfg
}

func getEnv(key, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val
}

// Helper to get int envs
func getEnvInt(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		log.Printf("Invalid int for %s: %v, using default %d", key, err, defaultVal)
		return defaultVal
	}
	return i
}
