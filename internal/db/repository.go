package db

import (
	"log"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4/pgxpool"
)

type DBRepository struct {
	pool   *pgxpool.Pool
	logger *log.Logger
	psql squirrel.StatementBuilderType
}

func NewRepository(pool *pgxpool.Pool, logger *log.Logger) *DBRepository {
	return &DBRepository{
		pool: pool, 
		logger: logger, 
		psql: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (repository *DBRepository) TearDown() {
	repository.pool.Close()
}
