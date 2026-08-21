package config

type Config struct {
	App AppConfig
	DB  DataBaseConfig
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
