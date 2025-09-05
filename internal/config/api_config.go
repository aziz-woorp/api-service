package config

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"

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
	MongoDB  string

	// RabbitMQ/Queue settings
	CeleryBrokerURL    string
	RabbitMQURL        string
	RabbitMQHost       string
	RabbitMQPort       int
	RabbitMQUser       string
	RabbitMQPassword   string
	RabbitMQVHost      string
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

	// Feature flags
	EnableClientChannelRouting   bool
	EnableConfigurableWorkflows  bool
}

func LoadConfig() *Config {
	// Load .env if present
	_ = godotenv.Load(".env")
	mongoURI := getEnv("MONGODB_URI", "mongodb://localhost:27017/fraiday-backend")
	
	cfg := &Config{
		// Application settings
		ProjectName: getEnv("PROJECT_NAME", "API Service"),
		Version:     getEnv("VERSION", "1.0.0"),
		AppPort:     getEnv("APP_PORT", "8000"),
		AppEnv:      getEnv("APP_ENV", "development"),
		GinMode:     getEnv("GIN_MODE", "debug"),
		LogLevel:    getEnv("LOG_LEVEL", "INFO"),

		// Database
		MongoURI: mongoURI,
		MongoDB:  extractDatabaseFromURI(mongoURI),

		// RabbitMQ/Queue settings
		CeleryBrokerURL:    getEnv("CELERY_BROKER_URL", ""),
		RabbitMQURL:        getEnv("RABBITMQ_URL", ""),
		RabbitMQHost:       getEnv("RABBITMQ_HOST", "localhost"),
		RabbitMQPort:       getEnvInt("RABBITMQ_PORT", 5672),
		RabbitMQUser:       getEnv("RABBITMQ_USER", "guest"),
		RabbitMQPassword:   getEnv("RABBITMQ_PASSWORD", "guest"),
		RabbitMQVHost:      getEnv("RABBITMQ_VHOST", "/"),
		CeleryDefaultQueue: getEnv("CELERY_DEFAULT_QUEUE", "chat_workflow"),
		CeleryEventsQueue:  getEnv("CELERY_EVENTS_QUEUE", "events"),

		// External services
		SlackAIServiceURL:       getEnv("SLACK_AI_SERVICE_URL", ""),
		SlackAIToken:            getEnv("SLACK_AI_TOKEN", ""),
		SlackAIServiceWorkflowID: getEnv("SLACK_AI_SERVICE_WORKFLOW_ID", ""),
		AIServiceURL:            getEnv("SLACK_AI_SERVICE_URL", ""),
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

		// Feature flags
		EnableClientChannelRouting:  getEnvBool("ENABLE_CLIENT_CHANNEL_ROUTING", false),
		EnableConfigurableWorkflows: getEnvBool("ENABLE_CONFIGURABLE_WORKFLOWS", false),
	}

	return cfg
}

// GetRabbitMQURL generates RabbitMQ URL from components if RABBITMQ_URL is not provided
func (c *Config) GetRabbitMQURL() string {
	// Use CeleryBrokerURL if available (for compatibility with Python backend)
	if c.CeleryBrokerURL != "" {
		return c.CeleryBrokerURL
	}
	if c.RabbitMQURL != "" {
		return c.RabbitMQURL
	}
	return fmt.Sprintf("amqp://%s:%s@%s:%d%s", c.RabbitMQUser, c.RabbitMQPassword, c.RabbitMQHost, c.RabbitMQPort, c.RabbitMQVHost)
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

// Helper to get bool envs
func getEnvBool(key string, defaultVal bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	b, err := strconv.ParseBool(val)
	if err != nil {
		log.Printf("Invalid bool for %s: %v, using default %t", key, err, defaultVal)
		return defaultVal
	}
	return b
}

// extractDatabaseFromURI extracts the database name from a MongoDB URI
func extractDatabaseFromURI(mongoURI string) string {
	// First check if there's an explicit MONGODB_DB environment variable
	if dbName := getEnv("MONGODB_DB", ""); dbName != "" {
		return dbName
	}

	// Parse the URI to extract database name
	parsedURI, err := url.Parse(mongoURI)
	if err != nil {
		log.Printf("Error parsing MongoDB URI: %v, using default database", err)
		return "fraiday-backend" // fallback
	}

	// Remove leading slash from path
	path := strings.TrimPrefix(parsedURI.Path, "/")
	
	// Remove query parameters if any (e.g., ?authSource=admin)
	if idx := strings.Index(path, "?"); idx != -1 {
		path = path[:idx]
	}

	// If path is empty or just "/", use default
	if path == "" {
		return "fraiday-backend"
	}

	return path
}
