package pgutils

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// Пул соединений с БД
var Pool *pgxpool.Pool

// Создание нового пула
func NewPool(cfg *DatabaseConfig) (*pgxpool.Pool, error) {
	return pgxpool.Connect(context.Background(), GetConnectionStringWithDBName(cfg))
}

// Первоначальная инициализация пула соединений
func InitPool(cfg *DatabaseConfig) error {
	var err error

	if Pool, err = NewPool(cfg); err != nil {
		log.Fatal("Can't init connection pgxpool with ", cfg.String(), "; Error: ", err)
	}
	return err
}

func ClosePool() {
	if Pool != nil {
		Pool.Close()
	}
}

// Структура ответа
type SimpleResultRespond struct {
	Result string `json:"result"`
}

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

// ------------------------------------------------------------------------------------------------
// Подключиться к БД используя конфигурационную структуру
func OpenDB(cfg *DatabaseConfig) (*pgx.Conn, error) {
	// Подключемся к базе (или с указанным именем БД или со стандартным postgress)
	connectionString := ""
	if len(cfg.DBname) > 0 {
		connectionString = GetConnectionStringWithDBName(cfg)
	} else {
		connectionString = GetConnectionString(cfg)
	}

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

// ------------------------------------------------------------------------------------------------
// Подключиться к БД используя строку connectionString
func OpenDBByString(connectionString string) (*pgx.Conn, error) {
	// Подключемся к базе
	conn, err := pgx.Connect(context.Background(), connectionString)
	if err != nil {
		fmt.Print(err)
		return conn, err
	}

	return conn, err
}

// ------------------------------------------------------------------------------------------------
func CreateDB(conn *pgx.Conn, dbname string) error {
	//создание новой БД
	log.Printf("Create database: %s", dbname)

	if err := conn.QueryRow(context.Background(), fmt.Sprintf("SELECT FROM pg_database WHERE datname = '%s'", dbname)).Scan(); err == pgx.ErrNoRows {
		_, err := conn.Exec(context.Background(), fmt.Sprintf(`CREATE DATABASE %s`, dbname))
		if err != nil {
			log.Println(err)
			return err
		}
	} else if err != nil {
		return err
	} else {
		log.Println("Database already exists!")
	}

	return nil
}

// ------------------------------------------------------------------------------------------------
func DropDB(conn *pgx.Conn, dbname string) error {
	log.Printf("Drop database %s", dbname)
	// Удаляем базу
	_, err := conn.Exec(context.Background(), `drop database if exists "`+dbname+`"`)
	if err != nil {
		return err
	}

	//проверка на существование БД
	var name string
	if err := conn.QueryRow(context.Background(), fmt.Sprintf(`SELECT datname FROM pg_database where datname = '%s'`, dbname)).Scan(&name); err != pgx.ErrNoRows || len(name) > 0 {
		log.Printf("drop database error or database not dropped")
		return err
	}
	return nil
}

// ------------------------------------------------------------------------------------------------
func DropSchema(conn *pgx.Conn, schema string) error {
	log.Printf("Drop schema %s with objects", schema)
	// Удаляем схему
	_, err := conn.Exec(context.Background(), `drop schema if exists "`+schema+`" cascade`)
	if err != nil {
		return err
	}

	//проверка на существование схемы
	var name string
	if err := conn.QueryRow(context.Background(), fmt.Sprintf(`SELECT schema_name FROM information_schema.schemata where schema_name = '%s'`, schema)).Scan(&name); err != pgx.ErrNoRows || len(name) > 0 {
		log.Printf("drop schema error or schema not dropped")
		return err
	}

	return nil
}

// ------------------------------------------------------------------------------------------------
func SetSchema(conn *pgx.Conn, name string) error {
	_, err := conn.Exec(context.Background(), `set schema '`+name+`'`)
	if err != nil {
		log.Print(err)
		return err
	}
	return nil
}

// ------------------------------------------------------------------------------------------------
func CreateSchema(conn *pgx.Conn, name string) error {
	log.Printf("Create schema: %s", name)
	//создание новой схемы
	_, err := conn.Exec(context.Background(), `create schema "`+name+`"`)

	return err
}

// ----------------------------------------------------------------------------
// Создание пользователя БД
func CreateSQLUser(conn *pgx.Conn, login string, pass string) error {
	//создание нового пользователя
	res := 0
	if err := conn.QueryRow(context.Background(), fmt.Sprintf(`SELECT 1 FROM pg_user WHERE "usename" = '%s'`, login)).Scan(&res); err == pgx.ErrNoRows {
		log.Printf("Create user: %s:%s", login, pass)
		_, err := conn.Exec(context.Background(), fmt.Sprintf(`CREATE USER %s WITH PASSWORD '%s' SUPERUSER`, login, pass))
		if err != nil {
			return err
		}
	}
	return nil
}

// ----------------------------------------------------------------------------
// Удаление пользователя БД
func DropSQLUser(conn *pgx.Conn, login string) error {
	//создание новой схемы
	log.Printf("Drop user: %s", login)
	_, err := conn.Exec(context.Background(), fmt.Sprintf("DROP USER %s", login))
	if err != nil {
		log.Print(err)
		return err
	}
	return nil
}

//------------------------------------------------------------------------------------------------
// Подготовка БД:
// 1. Удаление БД/Схемы указанной в конфигурации (опционально)
// 2. Создание БД указанной в конфигурации
// 3. Создание пользователя userap для внутренних доверенных сервисов
// 4. Создание указанной схемы

type Options struct {
	DropDB     bool
	DropSchema bool
}

func PrepareDBEnvironment(cfg *DatabaseConfig, opt Options) (*pgx.Conn, error) {
	log.Printf("Preparing database: %s:%d; %s:%s, Name: %s, Scheme: %s", cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBname, cfg.Schema)
	// Подсоединяемся чтобы удалить указанную в конфиге БД

	//------------------------------------------
	{ // Создаем БД указанную в конфиге - подключаемся без указания имени БД
		conn, err := OpenDB(&DatabaseConfig{Host: cfg.Host, Port: cfg.Port, User: cfg.User, Password: cfg.Password})
		if err != nil {
			return nil, err
		}

		if opt.DropDB {
			// Дропнем в начале
			err = DropDB(conn, cfg.DBname)
			if err != nil {
				conn.Close(context.Background())
				return conn, err
			}
		}

		if err = CreateDB(conn, cfg.DBname); err != nil {
			conn.Close(context.Background())
			return conn, err
		}
		conn.Close(context.Background())
	}
	///-----------
	{ // Подключаемся к созданной БД
		conn, err := OpenDB(cfg)
		if err != nil {
			return nil, err
		}

		if opt.DropSchema {
			// Дропнем в начале
			err = DropSchema(conn, cfg.Schema)
			if err != nil {
				conn.Close(context.Background())
				return conn, err
			}
		}

		// Создаем пользователя для сервисов
		if err = CreateSQLUser(conn, ServiceUser, ServiceUserPass); err != nil {
			log.Print(err)
		}

		if len(cfg.Schema) > 0 {
			// Создаем схему
			if err = CreateSchema(conn, cfg.Schema); err != nil {
				log.Print(err)
			}

			//переходим на работу со схемой даже есои ее е создали (public)
			err = SetSchema(conn, cfg.Schema)
			if err != nil {
				return conn, err
			}
		}

		return conn, nil
	}
}

// ----------------------------------------------------------------------------
func DropDBEnvironment(cfg *DatabaseConfig) error {
	conn, err := OpenDB(&DatabaseConfig{Host: cfg.Host, Port: cfg.Port, User: cfg.User, Password: cfg.Password})
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())
	// Дропнем в начале
	if err = DropDB(conn, cfg.DBname); err != nil {
		return err
	}
	// не уничтожаем чтобы после выполнения тестов база была рабочая
	//if err = DropSQLUser(conn, ServiceUser); err != nil {
	//	return err
	//}
	return err
}
