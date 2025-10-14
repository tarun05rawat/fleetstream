package config

import (
	"database/sql"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
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
	URL      string
	Host     string
	Port     int
	Name     string
	User     string
	Password string
	SSLMode  string
}

// KafkaConfig holds Kafka connection configuration
type KafkaConfig struct {
	Brokers    string
	GroupID    string
	Topics     []string
	AutoOffset string
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
			URL:      os.Getenv("DATABASE_URL"), // full connection string
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
// Prefers DATABASE_URL env var, otherwise builds from components
func (c *Config) GetDatabaseURL() string {
	if c.Database.URL != "" {
		// Force IPv4 for Supabase hostnames
		return forceIPv4(c.Database.URL)
	}
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host, c.Database.Port, c.Database.User,
		c.Database.Password, c.Database.Name, c.Database.SSLMode)
}

// forceIPv4 rewrites the hostname to an IPv4 address if possible
func forceIPv4(dsn string) string {
	if !(strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://")) {
    return dsn
	}


	parts := strings.Split(dsn, "@")
	if len(parts) != 2 {
		return dsn
	}

	creds := parts[0] // postgres://user:pass
	hostPart := parts[1]

	// Split host:port/dbname
	hostParts := strings.SplitN(hostPart, "/", 2)
	if len(hostParts) < 1 {
		return dsn
	}

	hostPort := hostParts[0]
	rest := ""
	if len(hostParts) > 1 {
		rest = "/" + hostParts[1]
	}

	// Separate host and port
	host, port, err := net.SplitHostPort(hostPort)
	if err != nil {
		// if no port given, assume 5432
		host = hostPort
		port = "5432"
	}

	// Lookup IPv4
	ips, err := net.LookupIP(host)
	if err != nil {
		return dsn
	}

	var ipv4 string
	for _, ip := range ips {
		if ip.To4() != nil {
			ipv4 = ip.String()
			break
		}
	}

	if ipv4 == "" {
		// fallback
		return dsn
	}

	schemeAndCreds := strings.SplitN(creds, "://", 2)
	if len(schemeAndCreds) != 2 {
			return dsn
	}
	scheme := schemeAndCreds[0]        // "postgres" or "postgresql"
	userPass := schemeAndCreds[1]      // "user:pass"
	return fmt.Sprintf("%s://%s@%s:%s%s", scheme, userPass, ipv4, port, rest)
}

// getEnvOrDefault returns environment variable value or default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Simple connectivity test
func TestDBConnection(dsn string) error {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}
	defer db.Close()
	return db.Ping()
}
