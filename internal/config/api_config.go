package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	// Application settings
	ProjectName string
	Version     string
	AppPort     string
	AppEnv      string
	GinMode     string
	LogLevel    string

	// Database
	MongoURI string

	// Celery/Queue settings
	CeleryBrokerURL    string
	CeleryDefaultQueue string
	CeleryEventsQueue  string

	// External services
	SlackAIServiceURL       string
	SlackAIToken            string
	SlackAIServiceWorkflowID string
	AIServiceURL            string
	EncryptionKey           string
	AdminAPIKey             string

	// AWS Bedrock
	AWSBedrockAccessKeyID     string
	AWSBedrockSecretAccessKey string
	AWSBedrockRegion          string
	AWSBedrockRuntime         string

	// Redis
	RedisHost     string
	RedisPort     int
	RedisDB       int
	RedisPassword string
}

func LoadConfig() *Config {
	// Load .env if present
	_ = godotenv.Load(".env")
	cfg := &Config{
		// Application settings
		ProjectName: getEnv("PROJECT_NAME", "AI Service Backend"),
		Version:     getEnv("VERSION", "0.0.1"),
		AppPort:     getEnv("APP_PORT", "8080"),
		AppEnv:      getEnv("APP_ENV", "development"),
		GinMode:     getEnv("GIN_MODE", "debug"),
		LogLevel:    getEnv("LOG_LEVEL", "INFO"),

		// Database
		MongoURI: getEnv("MONGODB_URI", "mongodb://localhost:27017/api_service_dev"),

		// Celery/Queue settings
		CeleryBrokerURL:    getEnv("CELERY_BROKER_URL", ""),
		CeleryDefaultQueue: getEnv("CELERY_DEFAULT_QUEUE", "chat_workflow"),
		CeleryEventsQueue:  getEnv("CELERY_EVENTS_QUEUE", "events"),

		// External services
		SlackAIServiceURL:       getEnv("SLACK_AI_SERVICE_URL", ""),
		SlackAIToken:            getEnv("SLACK_AI_TOKEN", ""),
		SlackAIServiceWorkflowID: getEnv("SLACK_AI_SERVICE_WORKFLOW_ID", ""),
		AIServiceURL:            getEnv("AI_SERVICE_URL", ""),
		EncryptionKey:           getEnv("ENCRYPTION_KEY", ""),
		AdminAPIKey:             getEnv("ADMIN_API_KEY", ""),

		// AWS Bedrock
		AWSBedrockAccessKeyID:     getEnv("AWS_BEDROCK_ACCESS_KEY_ID", ""),
		AWSBedrockSecretAccessKey: getEnv("AWS_BEDROCK_SECRET_ACCESS_KEY", ""),
		AWSBedrockRegion:          getEnv("AWS_BEDROCK_REGION", ""),
		AWSBedrockRuntime:         getEnv("AWS_BEDROCK_RUNTIME", "bedrock-runtime"),

		// Redis
		RedisHost:     getEnv("REDIS_HOST", ""),
		RedisPort:     getEnvInt("REDIS_PORT", 6379),
		RedisDB:       getEnvInt("REDIS_DB", 0),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
	}

	return cfg
}

// GetRedisURL generates Redis URL from components if CELERY_BROKER_URL is not provided
func (c *Config) GetRedisURL() string {
	if c.CeleryBrokerURL != "" {
		return c.CeleryBrokerURL
	}
	if c.RedisPassword != "" {
		return fmt.Sprintf("rediss://default:%s@%s:%d/%d", c.RedisPassword, c.RedisHost, c.RedisPort, c.RedisDB)
	}
	return fmt.Sprintf("redis://%s:%d/%d", c.RedisHost, c.RedisPort, c.RedisDB)
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
