package pgutils

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

type Querier interface {
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}

func ExecuteSqlScripts(tx pgx.Tx, scriptsPath string) error {
	// Выполним создание таблиц
	if files, err := os.ReadDir(scriptsPath); err == nil {
		for _, v := range files {
			if !v.IsDir() && strings.Contains(v.Name(), ".sql") {
				log.Printf("Executing: %s\n", v.Name())
				// Прочитаем файл создания БД этой версии и выполняем его
				if execSql, err := os.ReadFile(scriptsPath + "/" + v.Name()); err == nil {
					if _, err = tx.Exec(context.Background(), string(execSql)); err != nil {
						return err
					}
				}
			}
		}

		if len(files) == 0 {
			log.Print("No scripts found")
		}

		return err
	}
	return nil
}
