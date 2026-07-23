package config

import "os"

// Config holds all runtime configuration for the data-service.
type Config struct {
	DatabaseURL string
	GRPCPort    string
}

// Load reads configuration from environment variables, falling back to
// sensible defaults for local development.
func Load() *Config {
	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://fund_app:pass@localhost:5432/fund_db"),
		GRPCPort:    getEnv("GRPC_PORT", "9090"),
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
