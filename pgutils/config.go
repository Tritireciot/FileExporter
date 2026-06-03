package pgutils

import (
	"fmt"
	"log"
)

// Конфиг подключения к БД
type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBname   string `json:"dbname"`
	Schema   string `json:"schema"`
}

func (cfg *DatabaseConfig) String() string {
	return fmt.Sprintf("DataBase configuration:\n\tHost: %s\n\tPort: %d\n\tUsername: %s\n\tPassword: %s\n\tDBname: %s\n\tSchema: %s\n",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.DBname,
		cfg.Schema)
}

func (cfg *DatabaseConfig) Print() {
	log.Print(cfg.String())
}