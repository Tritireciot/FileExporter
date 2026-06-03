package db

import (
	"PrintServer/pgutils"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

func SetupDB(ctx context.Context, cfg *pgutils.DatabaseConfig, logger *log.Logger) (*DBRepository, error) {
	dbpool, err := pgxpool.Connect(ctx, pgutils.GetConnectionStringWithDBName(cfg))
	
	if err != nil {
		logger.Println("Unable to connect to database:", err.Error())
		dbpool.Close()
		return nil, err
	}
	repo := NewRepository(dbpool, logger)

	logger.Println("Check Connection")

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := repo.pool.Ping(ctx); err != nil {
		dbpool.Close()
		return nil, err
	}

	logger.Println("Check Fill")

	isFilled, err := repo.isDBFilled(ctx)

	if !isFilled {
		logger.Println("Fill DB")
		if err := repo.createFilledTables(ctx); err != nil {
			return nil, err
		}
	}
	logger.Println("DB READY")

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
	if err := repository.createTagsTables(ctx); err != nil {
		repository.pool.Close()
		return err
	}
	repository.logger.Println("Created Tag Table")

	if err := repository.createTemplatesTables(ctx); err != nil {
		repository.pool.Close()
		return err
	}
	repository.logger.Println("Created Template Table")

	if err := repository.fillTemplatesTable(ctx, os.Getenv("TEMPLATES_PATH")); err != nil {
		repository.pool.Close()
		return err
	}
	repository.logger.Println("Filled Template Table")

	if err := repository.fillTagsTable(ctx, os.Getenv("TAGS_FILEPATH")); err != nil {
		repository.pool.Close()
		return err
	}
	repository.logger.Println("Filled Tag Table")

	return nil

}
