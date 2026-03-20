package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database PostgresConfig
	Redis    RedisConfig
	Swagger  SwaggerConfig
}

type ServerConfig struct {
	Port        string
	Environment string
}

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
	CacheTTL time.Duration
}

type SwaggerConfig struct {
	Enabled bool
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found, using environment variables")
	}

	cacheTTL, err := time.ParseDuration(getEnv("CACHE_TTL", "5m"))
	if err != nil {
		cacheTTL = 5 * time.Minute
	}

	swaggerEnabled := getEnv("SWAGGER_ENABLED", "true") == "true"

	config := &Config{
		Server: ServerConfig{
			Port:        getEnv("SERVER_PORT", "8080"),
			Environment: getEnv("ENVIRONMENT", "development"),
		},
		Database: PostgresConfig{
			Host:     getEnv("POSTGRES_HOST", "localhost"),
			Port:     getEnv("POSTGRES_PORT", "5432"),
			User:     getEnv("POSTGRES_USER", "booksapi"),
			Password: getEnv("POSTGRES_PASSWORD", "booksapi_password"),
			Database: getEnv("POSTGRES_DB", "booksapi"),
			SSLMode:  getEnv("POSTGRES_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       0,
			CacheTTL: cacheTTL,
		},
		Swagger: SwaggerConfig{
			Enabled: swaggerEnabled,
		},
	}

	return config, nil
}

func (c *PostgresConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.Database, c.SSLMode)
}

func (c *RedisConfig) Address() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
