package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4"
)

type ShortElement struct {
	Element_id   int    `json:"element_id"`
	Element_name string `json:"element_name"`
	IsSingle bool `json:"is_single"`
}

func (repository *DBRepository) GetAllElements(ctx context.Context, element DBModel, subsystem string, isActive bool) (*[]ShortElement, error) {
	var sqr_query squirrel.SelectBuilder
	if _, ok := element.(*Template); ok {
		sqr_query = repository.psql.Select(Columns.ID, Columns.Name, Columns.IsSingle)
	} else {
		sqr_query = repository.psql.Select(Columns.ID, Columns.Name)
	}
	sqr_query = sqr_query.From(element.GetTable())

	if isActive {
		sqr_query = sqr_query.Where(Columns.IsActive)
	}
	if ok := ParseSubsystem(subsystem); ok {
		sqr_query = sqr_query.Where(squirrel.Eq{Columns.Subsystem: subsystem})
	}
	query, args, err := sqr_query.ToSql()
	if err != nil {
		return nil, err
	}
	var elements []ShortElement

	rows, err := repository.pool.Query(ctx, query, args...)
	if err != nil {
		rows.Close()
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		short_element := ShortElement{}
		if _, ok := element.(*Template); ok {
			err = rows.Scan(&short_element.Element_id, &short_element.Element_name, &short_element.IsSingle)
		} else {
			err = rows.Scan(&short_element.Element_id, &short_element.Element_name)
		}
		if err != nil {
			return nil, err
		}
		elements = append(elements, short_element)
	}

	return &elements, nil
}

func (repository *DBRepository) GetElement(
	ctx context.Context,
	element DBModel,
	column string,
) error {

	var field any

	if column == Columns.Name {
		field = element.GetName()
	} else {
		field = element.GetID()
	}
	query, args, err := repository.psql.Select("*").From(element.GetTable()).Where(squirrel.Eq{column: field}).ToSql()
	if err != nil {
		return err
	}
	err = repository.pool.
		QueryRow(ctx, query, args...).
		Scan(element.GetColumns()...)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrElementNotFound
		}
		return err
	}

	return nil
}

func (repository *DBRepository) DeleteElement(
	ctx context.Context,
	element DBModel,
	column string,
) error {
	var field any
	if column == Columns.Name {
		field = element.GetName()
	} else {
		field = element.GetID()
	}
	query, args, err := repository.psql.Delete(element.GetTable()).Where(squirrel.Eq{column: field}).ToSql()
	if err != nil {
		return err
	}
	cmd, err := repository.pool.Exec(ctx, query, args...)

	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return ErrElementNotFound
	}

	return nil
}

func (repository *DBRepository) AddTemplate(ctx context.Context, template *Template) error {
	sqr_query := repository.psql.Insert(Tables.Templates)
	if template.GetID() != 0 {
		sqr_query = sqr_query.Columns(Columns.ID, Columns.Name, Columns.Content, Columns.Subsystem, Columns.IsActive, Columns.RenderData, Columns.IsSingle).
		Values(template.GetColumns()...).Suffix(
			fmt.Sprintf(`
				ON CONFLICT (%s) DO UPDATE
				SET 
					%s = EXCLUDED.%s,
					%s = EXCLUDED.%s,
					%s = EXCLUDED.%s,
					%s = EXCLUDED.%s,
					%s = EXCLUDED.%s,
					%s = EXCLUDED.%s
				`,
				Columns.ID,
				Columns.Name, Columns.Name,
				Columns.Content, Columns.Content,
				Columns.Subsystem, Columns.Subsystem,
				Columns.IsActive, Columns.IsActive,
				Columns.RenderData, Columns.RenderData,
				Columns.IsSingle, Columns.IsSingle,
			),
		).Suffix(fmt.Sprintf("RETURNING %s;", Columns.ID))
	} else {
		sqr_query = sqr_query.Columns(Columns.Name, Columns.Content, Columns.Subsystem, Columns.IsActive, Columns.RenderData, Columns.IsSingle).
		Values(template.GetColumns()[1:]...).Suffix(fmt.Sprintf("RETURNING %s;", Columns.ID))
	}

	query, args, err := sqr_query.ToSql()
	if err != nil {
		return err
	}
	err = repository.pool.QueryRow(ctx, query, args...).Scan(&template.ID)
	return err
}

func (repository *DBRepository) AddTag(ctx context.Context, tag *Tag) error {
	query, args, err := repository.psql.Insert(Tables.Tags).
		Columns(Columns.Name, Columns.Description, Columns.Subsystem, Columns.Alias).
		Values(tag.Name, tag.Description, tag.Subsystem, tag.Alias).Suffix(
		fmt.Sprintf(`
			ON CONFLICT (%s) DO UPDATE
			SET 
				%s = EXCLUDED.%s,
				%s = EXCLUDED.%s,
				%s = EXCLUDED.%s
			`,
			Columns.Name,
			Columns.Description, Columns.Description,
			Columns.Subsystem, Columns.Subsystem,
			Columns.Alias, Columns.Alias,
		),
	).Suffix(fmt.Sprintf("RETURNING %s;", Columns.ID)).ToSql()
	if err != nil {
		return err
	}

	err = repository.pool.QueryRow(ctx, query, args...).Scan(&tag.ID)
	return err
}

func (repository *DBRepository) createTagsTables(ctx context.Context) error {
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		%s SERIAL PRIMARY KEY,
		%s TEXT NOT NULL UNIQUE,
		%s TEXT NOT NULL,
		%s TEXT NOT NULL,
		%s TEXT NOT NULL,
		%s BOOLEAN DEFAULT TRUE);
	`,
		Tables.Tags,
		Columns.ID,
		Columns.Name,
		Columns.Description,
		Columns.Subsystem,
		Columns.Alias,
		Columns.IsActive)
	_, err := repository.pool.Exec(ctx, query)
	return err
}

func (repository *DBRepository) createTemplatesTables(ctx context.Context) error {
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		%s SERIAL PRIMARY KEY,
		%s TEXT NOT NULL,
		%s TEXT NOT NULL,
		%s TEXT NOT NULL,
		%s BOOLEAN DEFAULT TRUE,
		%s JSONB DEFAULT '{}',
		%s BOOLEAN DEFAULT FALSE);
	`,
		Tables.Templates,
		Columns.ID,
		Columns.Name,
		Columns.Content,
		Columns.Subsystem,
		Columns.IsActive,
		Columns.RenderData,
		Columns.IsSingle)
	_, err := repository.pool.Exec(ctx, query)
	return err
}
