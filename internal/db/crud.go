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
}

func (repository *DBRepository) GetAllElements(ctx context.Context, element DBModel) (*[]ShortElement, error) {
	query, _, err:= repository.psql.Select(Columns.ID, Columns.Name).From(element.GetTable()).ToSql()
	if err != nil {
		return nil, err
	}
	var elements []ShortElement

	rows, err := repository.pool.Query(ctx, query)
	if err != nil {
		rows.Close()
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		short_element := ShortElement{}
		err := rows.Scan(&short_element.Element_id, &short_element.Element_name)
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
			return ErrTemplateNotFound
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
	query, args, err := repository.psql.Insert(Tables.Templates).
	Columns(Columns.Name, Columns.Content).
	Values(template.Name, template.Content).Suffix(
		fmt.Sprintf(`
			ON CONFLICT (%s) DO UPDATE
			SET %s = EXCLUDED.%s
			`, 
			Columns.Name,
			Columns.Content, Columns.Content,
		),
	).Suffix(fmt.Sprintf("RETURNING %s;", Columns.ID)).ToSql()
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

	_, err = repository.pool.Exec(ctx, query, args...)
	return err
}

func (repository *DBRepository) createTagsTables(ctx context.Context) error {
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		%s SERIAL PRIMARY KEY,
		%s TEXT NOT NULL UNIQUE,
		%s TEXT NOT NULL,
		%s TEXT NOT NULL,
		%s TEXT NOT NULL);
	`,
		Tables.Tags,
		Columns.ID,
		Columns.Name,
		Columns.Description,
		Columns.Subsystem,
		Columns.Alias)
	_, err := repository.pool.Exec(ctx, query)
	return err
}

func (repository *DBRepository) createTemplatesTables(ctx context.Context) error {
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		%s SERIAL PRIMARY KEY,
		%s TEXT NOT NULL UNIQUE,
		%s TEXT NOT NULL);
	`,
		Tables.Templates,
		Columns.ID,
		Columns.Name,
		Columns.Content)
	_, err := repository.pool.Exec(ctx, query)
	return err
}
