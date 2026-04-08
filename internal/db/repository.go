package db

import (
	"log"

	"github.com/jackc/pgx/v4/pgxpool"
)

type DBRepository struct {
	pool   *pgxpool.Pool
	logger *log.Logger
}

func NewRepository(pool *pgxpool.Pool, logger *log.Logger) *DBRepository {
	return &DBRepository{pool: pool, logger: logger}
}
