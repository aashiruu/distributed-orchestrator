package config

import "os"

type Config struct {
	DBURL    string
	AMQPURL  string
	RedisURL string
	APIPort  string
}

func Load() *Config {
	return &Config{
		DBURL:    getEnv("DATABASE_URL", "postgres://orchestrator_user:orchestrator_password@localhost:5432/job_orchestrator?sslmode=disable"),
		AMQPURL:  getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		RedisURL: getEnv("REDIS_URL", "localhost:6379"),
		APIPort:  getEnv("API_PORT", "8080"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
