package config

import (
	"os"
)

func getEnv(key, default_value string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return default_value
}

func LoadConfig() Config {
	return Config{
		App: AppConfig{
			Host: "0.0.0.0",
			Port: getEnv("HTTP_PORT", "8153"),
		},
	}
}
