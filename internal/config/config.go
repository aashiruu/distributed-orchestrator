package config

import (
	"os"
)

type Config struct {
	DBURL    string
	AMQPURL  string
	RedisURL string
	APIPort  string
}

func Load() *Config {
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		dbURL = "postgres://orchestrator_user:orchestrator_password@localhost:5432/job_orchestrator?sslmode=disable"
	}

	amqpURL := os.Getenv("AMQP_URL")
	if amqpURL == "" {
		amqpURL = "amqp://guest:guest@localhost:5672/"
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379/0"
	}

	apiPort := os.Getenv("API_PORT")
	if apiPort == "" {
		apiPort = "8080"
	}

	return &Config{
		DBURL:    dbURL,
		AMQPURL:  amqpURL,
		RedisURL: redisURL,
		APIPort:  apiPort,
	}
}
