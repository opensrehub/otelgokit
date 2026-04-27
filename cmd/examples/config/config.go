package config

import (
	"os"
	"strconv"
)

// Config holds service configuration
type Config struct {
	ServiceName  string
	Environment  string
	CollectorURL string
	Port         int
}

// LoadConfig initializes service configuration from environment variables
func LoadConfig() Config {
	port, err := strconv.Atoi(getEnv("PORT", "3000"))
	if err != nil {
		port = 3000
	}
	return Config{
		ServiceName:  getEnv("SERVICE_NAME", "otel-service"),
		Environment:  getEnv("ENVIRONMENT", "development"),
		CollectorURL: getEnv("OTEL_COLLECTOR_URL", "localhost:4317"),
		Port:         port,
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
