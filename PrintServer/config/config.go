package config

type Config struct {
	App AppConfig
}

type AppConfig struct {
	Host string
	Port string
}

type DataBaseConfig struct {
	DB_URL string
}

func (cfg Config) ConfigureAppUrl() string {
	return ":" + cfg.App.Port
}
