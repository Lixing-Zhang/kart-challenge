package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds all configuration for the application
// Following 12-factor app principles, all config is loaded from environment variables
type Config struct {
	Server   ServerConfig
	Auth     AuthConfig
	Coupon   CouponConfig
	LogLevel string
}

type ServerConfig struct {
	Port            string
	Host            string
	ReadTimeout     int
	WriteTimeout    int
	ShutdownTimeout int
}

type AuthConfig struct {
	APIKeys []string // Valid API keys for authentication
}

type CouponConfig struct {
	File1URL string
	File2URL string
	File3URL string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port:            getEnv("PORT", "8080"),
			Host:            getEnv("HOST", "0.0.0.0"),
			ReadTimeout:     getEnvAsInt("READ_TIMEOUT", 15),
			WriteTimeout:    getEnvAsInt("WRITE_TIMEOUT", 15),
			ShutdownTimeout: getEnvAsInt("SHUTDOWN_TIMEOUT", 30),
		},
		Auth: AuthConfig{
			APIKeys: getEnvAsSlice("API_KEYS", []string{"apitest"}),
		},
		Coupon: CouponConfig{
			File1URL: getEnv("COUPON_FILE1_URL", "https://orderfoodonline-files.s3.ap-southeast-2.amazonaws.com/couponbase1.gz"),
			File2URL: getEnv("COUPON_FILE2_URL", "https://orderfoodonline-files.s3.ap-southeast-2.amazonaws.com/couponbase2.gz"),
			File3URL: getEnv("COUPON_FILE3_URL", "https://orderfoodonline-files.s3.ap-southeast-2.amazonaws.com/couponbase3.gz"),
		},
		LogLevel: getEnv("LOG_LEVEL", "info"),
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Server.Port == "" {
		return fmt.Errorf("PORT is required")
	}

	if len(c.Auth.APIKeys) == 0 {
		return fmt.Errorf("at least one API key must be configured")
	}

	validLogLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLogLevels[strings.ToLower(c.LogLevel)] {
		return fmt.Errorf("invalid log level: %s (must be debug, info, warn, or error)", c.LogLevel)
	}

	return nil
}

// Helper functions for reading environment variables

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	return strings.Split(valueStr, ",")
}
