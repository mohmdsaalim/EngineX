package config

import (
	"os"
	"time"
)

// Config holds all configuration for every service.
// Each service picks what it needs from this struct.
type Config struct {
	// Database
	PostgresDSN string

	// Redis
	RedisAddr     string
	RedisPassword string

	// Kafka
	KafkaBroker string

	// JWT
	JWTSecret      string
	JWTAccessTTL   time.Duration
	JWTRefreshTTL  time.Duration

	// Service ports
	GatewayPort      string
	AuthGRPCPort     string
	RiskGRPCPort     string
	EngineGRPCPort   string
	ExecutorGRPCPort string
	WSHubPort        string
}

// Load reads env vars and returns a Config.
// If an env var is missing it falls back to a safe default.
func Load() *Config {
	return &Config{
		PostgresDSN:      getEnv("POSTGRES_DSN", "postgres://engine_user:engine_pass@localhost:5432/engine_db?sslmode=disable"),
		RedisAddr:        getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:    getEnv("REDIS_PASSWORD", ""),
		KafkaBroker:      getEnv("KAFKA_BROKER", "localhost:9092"),
		JWTSecret:        getEnv("JWT_SECRET", "dev-secret"),
		JWTAccessTTL:     getDuration("JWT_ACCESS_TTL", 15*time.Minute),
		JWTRefreshTTL:    getDuration("JWT_REFRESH_TTL", 168*time.Hour),
		GatewayPort:      getEnv("GATEWAY_PORT", ":8080"),
		AuthGRPCPort:     getEnv("AUTH_GRPC_PORT", ":9091"),
		RiskGRPCPort:     getEnv("RISK_GRPC_PORT", ":9092"),
		EngineGRPCPort:   getEnv("ENGINE_GRPC_PORT", ":9093"),
		ExecutorGRPCPort: getEnv("EXECUTOR_GRPC_PORT", ":9094"),
		WSHubPort:        getEnv("WSHUB_PORT", ":8081"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getDuration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}