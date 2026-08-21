package db

import (
	"AutoplayX/pgutils"
	logging "PrintServer/agent"
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

func connectWithRetry(ctx context.Context, dbName string) (*pgxpool.Pool, error) {
	var dbpool *pgxpool.Pool
	var err error
	for i := 0; i < 5; i++ {
		dbpool, err = pgxpool.Connect(ctx, dbName)
		if err == nil {
			if err = dbpool.Ping(ctx); err == nil {
				return dbpool, nil
			}
		}
		time.Sleep(2 * time.Second)
	}
	return nil, fmt.Errorf("база данных так и не ответила: %w", err)
}

func SetupDB(ctx context.Context, cfg *pgutils.DatabaseConfig) (*DBRepository, error) {
	dbpool, err := connectWithRetry(ctx, pgutils.GetConnectionStringWithDBName(cfg))

	if err != nil {
		logging.Agent.AddSimpleError("Подключение к БД", "Не удалось подключиться к БД: "+err.Error())

		return nil, err
	}

	repo := NewRepository(dbpool)

	logging.Agent.AddSimpleInfo("Проверка БД", "Подключение к БД...")

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := repo.pool.Ping(ctx); err != nil {
		dbpool.Close()
		return nil, err
	}

	logging.Agent.AddSimpleInfo("Проверка БД", "Проверка заполнения")

	isFilled, err := repo.isDBFilled(ctx)

	if !isFilled || err != nil {
		logging.Agent.AddSimpleInfo("Проверка БД", "Заполнение БД базовыми данными")
		if err := repo.createFilledTables(ctx); err != nil {
			return nil, err
		}
	}
	logging.Agent.AddSimpleInfo("Проверка БД", "БД готова к работе!")

	return repo, nil
}

func (repository *DBRepository) isDBFilled(ctx context.Context) (bool, error) {
	tableExistsQuery, args, err := repository.psql.Select("COUNT(*) > 0").
		From("information_schema.tables").
		Where(squirrel.Eq{
			"table_schema": "print",
			"table_name":   Tables.Templates, // ваша переменная "templates"
		}).
		ToSql()

	if err != nil {
		return false, fmt.Errorf("ошибка сборки SQL (проверка таблицы): %w", err)
	}

	var tableExists bool
	if err := repository.pool.QueryRow(ctx, tableExistsQuery, args...).Scan(&tableExists); err != nil {
		return false, err
	}

	if !tableExists {
		logging.Agent.AddSimpleInfo("Проверка БД", "Таблица templates еще не создана.")
		return false, nil
	}

	dataExistsQuery, args, err := repository.psql.Select("1").
		From("print." + Tables.Templates).
		Limit(1).
		ToSql()

	if err != nil {
		return false, fmt.Errorf("ошибка сборки SQL (проверка данных): %w", err)
	}
	var hasData int
	err = repository.pool.QueryRow(ctx, dataExistsQuery, args...).Scan(&hasData)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (repository *DBRepository) createFilledTables(ctx context.Context) error {

	if err := repository.createSubsystemsTables(ctx); err != nil {
		logging.Agent.AddSimpleInfo("Заполнение БД", "Не получилось создать Таблицу Подсистем "+err.Error())
		repository.pool.Close()
		return err
	}
	logging.Agent.AddSimpleInfo("Заполнение БД", "Таблица подсистем создана")

	if err := repository.createTagsTables(ctx); err != nil {
		logging.Agent.AddSimpleInfo("Заполнение БД", "Не получилось создать Таблицу Тэгов "+err.Error())
		repository.pool.Close()
		return err
	}
	logging.Agent.AddSimpleInfo("Заполнение БД", "Таблица тэгов создана")

	if err := repository.createTemplatesTables(ctx); err != nil {
		logging.Agent.AddSimpleInfo("Заполнение БД", "Не получилось создать Таблицу Шаблонов "+err.Error())
		repository.pool.Close()
		return err
	}
	logging.Agent.AddSimpleInfo("Заполнение БД", "Таблица шаблонов создана")

	repository.AddSubsystem(ctx, "news")
	repository.AddSubsystem(ctx, "plan")

	if templates_path := os.Getenv("TEMPLATES_PATH"); templates_path != "" {
		if _, err := os.Stat(templates_path); !os.IsNotExist(err) {
			if err := repository.fillTemplatesTable(ctx, templates_path); err != nil {
				logging.Agent.AddSimpleInfo("Заполнение БД", "Не заполнить шаблоны "+err.Error())
				repository.pool.Close()
				return err
			}
			logging.Agent.AddSimpleInfo("Заполнение БД", "Заполнение шаблонами по умолчанию")
		}
	}

	if tags_path := os.Getenv("TAGS_FILEPATH"); tags_path != "" {
		if _, err := os.Stat(tags_path); !os.IsNotExist(err) {
			if err := repository.fillTagsTable(ctx, tags_path); err != nil {
				logging.Agent.AddSimpleInfo("Заполнение БД", "Не заполнить теги"+err.Error())
				repository.pool.Close()
				return err
			}
			logging.Agent.AddSimpleInfo("Заполнение БД", "Заполнение тегами по умолчанию")
		}
	}

	return nil

}
