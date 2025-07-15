package config

import "os"

type Config struct {
	ListenAddr  string
	DatabaseURL string
	AgentURL    string
	AgentToken  string
}

func Load() Config {
	return Config{
		ListenAddr:  getEnv("WORKER_LISTEN_ADDR", ":8081"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://user:pass@localhost:5432/dbname?sslmode=disable"),
		AgentURL:    getEnv("AGENT_URL", "http://localhost:8080/v2/transcript"),
		AgentToken:  getEnv("AGENT_TOKEN", "http://localhost:8080/v2/transcript"),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
