package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Kafka    KafkaConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port         string
	AllowOrigins []string
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host     string
	Port     int
	Name     string
	User     string
	Password string
	SSLMode  string
}

// KafkaConfig holds Kafka connection configuration
type KafkaConfig struct {
	Brokers     string
	GroupID     string
	Topics      []string
	AutoOffset  string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	dbPort, err := strconv.Atoi(getEnvOrDefault("DB_PORT", "5432"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %v", err)
	}

	return &Config{
		Server: ServerConfig{
			Port: getEnvOrDefault("SERVER_PORT", "8080"),
			AllowOrigins: []string{
				getEnvOrDefault("FRONTEND_URL", "http://localhost:3000"),
				"http://localhost:3000",
			},
		},
		Database: DatabaseConfig{
			Host:     getEnvOrDefault("DB_HOST", "localhost"),
			Port:     dbPort,
			Name:     getEnvOrDefault("DB_NAME", "factoryflow"),
			User:     getEnvOrDefault("DB_USER", "factoryuser"),
			Password: getEnvOrDefault("DB_PASSWORD", "factorypass"),
			SSLMode:  getEnvOrDefault("DB_SSLMODE", "disable"),
		},
		Kafka: KafkaConfig{
			Brokers:    getEnvOrDefault("KAFKA_BROKERS", "localhost:9092"),
			GroupID:    getEnvOrDefault("KAFKA_GROUP_ID", "factoryflow-backend"),
			Topics:     []string{getEnvOrDefault("KAFKA_TOPIC", "line1.sensor")},
			AutoOffset: getEnvOrDefault("KAFKA_AUTO_OFFSET", "latest"),
		},
	}, nil
}

// GetDatabaseURL returns formatted database connection URL
func (c *Config) GetDatabaseURL() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host, c.Database.Port, c.Database.User,
		c.Database.Password, c.Database.Name, c.Database.SSLMode)
}

// getEnvOrDefault returns environment variable value or default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}