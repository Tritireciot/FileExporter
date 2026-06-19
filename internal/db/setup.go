package db

import (
	logging "PrintServer/agent"
	"PrintServer/pgutils"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

func SetupDB(ctx context.Context, cfg *pgutils.DatabaseConfig) (*DBRepository, error) {
	dbpool, err := pgxpool.Connect(ctx, pgutils.GetConnectionStringWithDBName(cfg))
	
	if err != nil {
		logging.Agent.AddSimpleError("Подключение к БД", "Не удалось подключиться к БД: " + err.Error())

		dbpool.Close()
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

	if !isFilled {
		logging.Agent.AddSimpleInfo("Проверка БД", "Заполнение БД базовыми данными")
		if err := repo.createFilledTables(ctx); err != nil {
			return nil, err
		}
	}
	logging.Agent.AddSimpleInfo("Проверка БД", "БД готова к работе!")

	return repo, nil
}

func (repository *DBRepository) isDBFilled(ctx context.Context) (bool, error) {
	var exists bool
	err := repository.pool.QueryRow(ctx, fmt.Sprintf(`
		SELECT EXISTS (
			SELECT FROM %s LIMIT 1
		)
	`, Tables.Templates)).Scan(&exists)

	return exists, err
}

func (repository *DBRepository) createFilledTables(ctx context.Context) error {

	if err := repository.createSubsystemsTables(ctx); err != nil {
		repository.pool.Close()
		return err
	}
	logging.Agent.AddSimpleInfo("Заполнение БД", "Таблица подсистем создана")

	if err := repository.createTagsTables(ctx); err != nil {
		repository.pool.Close()
		return err
	}
	logging.Agent.AddSimpleInfo("Заполнение БД", "Таблица тэгов создана")

	if err := repository.createTemplatesTables(ctx); err != nil {
		repository.pool.Close()
		return err
	}
	logging.Agent.AddSimpleInfo("Заполнение БД", "Таблица шаблонов создана")

	repository.AddSubsystem(ctx, "news", "http://mock_backend:8150/api/moc/news") // Temporary
	repository.AddSubsystem(ctx, "plan", "http://mock_backend:8150/api/moc/plan") // Temporary


	if err := repository.fillTemplatesTable(ctx, os.Getenv("TEMPLATES_PATH")); err != nil {
		repository.pool.Close()
		return err
	}
	logging.Agent.AddSimpleInfo("Заполнение БД", "Заполнение шаблонами по умолчанию")

	if err := repository.fillTagsTable(ctx, os.Getenv("TAGS_FILEPATH")); err != nil {
		repository.pool.Close()
		return err
	}
	logging.Agent.AddSimpleInfo("Заполнение БД", "Заполнение тегами по умолчанию")

	return nil

}
