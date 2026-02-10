package config

import "os"

func getEnv(key, default_value string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return default_value
}

func LoadConfig() Config {
	return Config{
		App: AppConfig{
			Host: getEnv("APP_HOST", "0.0.0.0"),
			Port: getEnv("APP_PORT", "8000"),
		},
		DB: DataBaseConfig{
			DB_URL: getEnv("DATABASE_URL", "postgres://user:pass@db:5432/backend-db"),
		},
	}
}