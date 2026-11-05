package config

import "os"

type Config struct {
	database DatabaseConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) Load() error {
	c.database = DatabaseConfig{
		Host:     getEnv("DB_HOST", "postgres"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "postgres"),
		Name:     getEnv("DB_NAME", "tools_ecg"),
	}

	return nil
}

func (c *Config) GetDatabaseConfig() DatabaseConfig {
	return c.database
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}