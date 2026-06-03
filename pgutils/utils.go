package pgutils

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v4"
)

func GetConnectionString(cfg *DatabaseConfig) string {
	s := fmt.Sprintf("host=%s port=%d user=%s dbname=postgres sslmode=disable application_name=InternalService", cfg.Host, cfg.Port, cfg.User)
	
	if len(cfg.Password) > 0 {
		s += fmt.Sprintf(" password=%s", cfg.Password)
	}
	return s
}

func GetConnectionStringWithDBName(cfg *DatabaseConfig) string {
	s := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable application_name=InternalService", cfg.Host, cfg.Port, cfg.User, cfg.DBname)
	if len(cfg.Password) > 0 {
		s += fmt.Sprintf(" password=%s", cfg.Password)
	}
	return s
}

// Пользователь для доступа к БД из внутренних сервисов
const ServiceUser = "userap"
const ServiceUserPass = "ap"

func OpenDB(cfg *DatabaseConfig) (*pgx.Conn, error) {
	// Подключемся к базе (или с указанным именем БД или со стандартным postgress)
	connectionString := ""
	if len(cfg.DBname) > 0 {
		connectionString = GetConnectionStringWithDBName(cfg)
	} else {
		connectionString = GetConnectionString(cfg)
	}

	log.Print("Connect string: ", connectionString)

	conn, err := pgx.Connect(context.Background(), connectionString)

	if err != nil {
		fmt.Print(err)
		return conn, err
	}

	return conn, err
}

// ------------------------------------------------------------------------------------------------
// Подключиться к БД используя конфигурационную структуру и переключиться на схему schema
func OpenDBWithSchema(cfg *DatabaseConfig, schema string) (*pgx.Conn, error) {
	// Подключемся к базе
	conn, err := OpenDB(cfg)
	if err != nil {
		fmt.Print(err)
		return conn, err
	}

	//переходим на работу с новой схемой
	err = SetSchema(conn, schema)

	return conn, err
}

func SetSchema(conn *pgx.Conn, name string) error {
	_, err := conn.Exec(context.Background(), `set schema '`+name+`'`)
	if err != nil {
		log.Print(err)
		return err
	}
	return nil
}