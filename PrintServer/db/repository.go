package db

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4/pgxpool"
)

type DBRepo interface {
	GetAllElements(ctx context.Context, element DBModel, subsystem string, isActive bool) (*[]ShortElement, error)
	GetElement(ctx context.Context, element DBModel, column string) error
	DeleteElement(ctx context.Context, element DBModel, column string) error
	AddTemplate(ctx context.Context, template *Template) error
	AddTag(ctx context.Context, tag *Tag) error
}

type DBRepository struct {
	pool *pgxpool.Pool
	psql squirrel.StatementBuilderType
}

func NewRepository(pool *pgxpool.Pool) *DBRepository {
	return &DBRepository{
		pool: pool,
		psql: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (repository *DBRepository) TearDown() {
	repository.pool.Close()
}
