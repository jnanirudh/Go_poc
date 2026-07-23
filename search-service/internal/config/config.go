package config

import "os"

// Config holds all runtime configuration for the search-service.
type Config struct {
	DatabaseURL     string
	GRPCPort        string
	DataServiceAddr string
}

// Load reads configuration from environment variables, falling back to
// sensible defaults for local development.
func Load() *Config {
	return &Config{
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://fund_app:pass@localhost:5432/fund_db"),
		GRPCPort:        getEnv("GRPC_PORT", "9091"),
		DataServiceAddr: getEnv("DATA_SERVICE_ADDR", "localhost:9090"),
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
