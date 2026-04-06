package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v4"
)

type ShortElement struct {
	Element_id int `json:"element_id"`
	Element_name string `json:"element_name"`
}

func (repository *DBRepository) GetAllElements(ctx context.Context, element DBModel) (*[]ShortElement, error) {
	query := fmt.Sprintf(
		"SELECT %s, %s FROM %s",
		Columns.ID, Columns.Name, element.getTable(),
	)

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

	query := fmt.Sprintf(
		"SELECT * FROM %s WHERE %s = $1",
		element.getTable(), column,
	)
	var field any
	
	if column == Columns.Name {
		field = element.getName()
	} else {
		field = element.getID()
	}
	err := repository.pool.
			QueryRow(ctx, query, field).
			Scan(element.getColumns()...)

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
	query := fmt.Sprintf(
		"DELETE FROM %s WHERE %s = $1",
		element.getTable(), column,
	)

	var field any
	if column == Columns.Name {
		field = element.getName()
	} else {
		field = element.getID()
	}
	cmd, err := repository.pool.Exec(ctx, query, field)

	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return ErrElementNotFound
	}

	return nil
}

func (repository *DBRepository) AddTemplate(ctx context.Context, template *Template) error {
	query := fmt.Sprintf(
		`INSERT INTO %s (%s, %s)
			VALUES ($1, $2)
			ON CONFLICT (%s) DO UPDATE
			SET %s = EXCLUDED.%s;
			`,
		Tables.Templates, Columns.Name, Columns.Content,
		Columns.Name,
		Columns.Content, Columns.Content,
	)
	_, err := repository.pool.Exec(ctx, query, template.Name, template.Content)
	return err
}

func (repository *DBRepository) AddTag(ctx context.Context, tag *Tag) error {
	query := fmt.Sprintf(
		`INSERT INTO %s (%s, %s, %s, %s)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (%s) DO UPDATE
			SET 
				%s = EXCLUDED.%s,
				%s = EXCLUDED.%s,
				%s = EXCLUDED.%s;
			`,
		Tables.Tags, Columns.Name, Columns.Description, Columns.Subsystem, Columns.Alias,
		Columns.Name,
		Columns.Description, Columns.Description,
		Columns.Subsystem, Columns.Subsystem,
		Columns.Alias, Columns.Alias,
	)
	_, err := repository.pool.Exec(ctx, query, tag.Name, tag.Description, tag.Subsystem, tag.Alias)
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
