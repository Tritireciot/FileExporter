package db

import (
	logging "PrintServer/agent"
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
	subsystem_id, err := repository.GetSubsystemId(ctx, subsystem)
	if subsystem_id == 0 {
		logging.Agent.AddSimpleError("Получение элементов", "Несуществующая подсистема: " + subsystem + err.Error())
		return nil, err
	}
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
	sqr_query = sqr_query.Where(squirrel.Eq{Columns.Subsystem: subsystem_id})
	query, args, err := sqr_query.ToSql()
	if err != nil {
		logging.Agent.AddSimpleError("Получение элементов", "Не удалось сформировать sql запрос: "+ err.Error())
		return nil, err
	}
	var elements []ShortElement

	rows, err := repository.pool.Query(ctx, query, args...)
	if err != nil {
		rows.Close()
		logging.Agent.AddSimpleError("Получение элементов", "Не удалось сформировать sql запрос: "+ err.Error())
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
			logging.Agent.AddSimpleError("Получение элементов", "Не удалось получить данные: "+ err.Error())
			return nil, err
		}
		elements = append(elements, short_element)
	}

	return &elements, nil
}

func (repository *DBRepository) GetRawElement(
	ctx context.Context,
	element DBModel,
	column string,
) error {
	query, args, err := repository.psql.Select(element.Columns()...).From(element.GetTable()).Where(squirrel.Eq{fmt.Sprintf("%s.%s", element.GetTable(), column): element.GetValueByColumn(column)}).ToSql()
	if err != nil {
		logging.Agent.AddSimpleError("Получение элемента", "Не удалось сформировать sql запрос: "+ err.Error())
		return err
	}
	err = repository.pool.
		QueryRow(ctx, query, args...).
		Scan(element.GetColumns()...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logging.Agent.AddSimpleError("Получение элемента", "Не найден: "+ err.Error())
			return ErrElementNotFound
		}
		logging.Agent.AddSimpleError("Получение элемента", "Не удалось выполнить sql запрос: "+ err.Error())
		return err
	}

	return nil
}

func (repository *DBRepository) GetElement(
	ctx context.Context,
	element DBModel,
	column string,
) error {

	var field any

	if column == Columns.ID {
		field = element.GetID()
	} else {
		field = element.GetName()
	}
	query, args, err := repository.psql.Select(element.Columns()...).From(element.GetTable()).
	LeftJoin(fmt.Sprintf("%s ON %s.%s = %s.%s", Tables.Subsystems, element.GetTable(), Columns.Subsystem, Tables.Subsystems, Columns.ID)).
	Where(squirrel.Eq{fmt.Sprintf("%s.%s", element.GetTable(), column): field}).ToSql()
	if err != nil {
		logging.Agent.AddSimpleError("Получение элемента", "Не удалось сформировать sql запрос: "+ err.Error())
		return err
	}
	err = repository.pool.
		QueryRow(ctx, query, args...).
		Scan(element.GetColumns()...)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logging.Agent.AddSimpleError("Получение элемента", "Не найден: "+ err.Error())
			return ErrElementNotFound
		}
		logging.Agent.AddSimpleError("Получение элемента", "Не удалось выполнить sql запрос: "+ err.Error())
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
		logging.Agent.AddSimpleError("Удаление элемента", "Не удалось сформировать sql запрос: "+ err.Error())
		return err
	}
	cmd, err := repository.pool.Exec(ctx, query, args...)

	if err != nil {
		logging.Agent.AddSimpleError("Удаление элемента", "Не удалось выполнить sql запрос: "+ err.Error())
		return err
	}

	if cmd.RowsAffected() == 0 {
		return ErrElementNotFound
	}

	return nil
}

func (repository *DBRepository) GetSubsystemId(ctx context.Context, subsystem string) (subsystem_id int, err error) {
	query, args, err := repository.psql.Select(Columns.ID).From(Tables.Subsystems).Where(squirrel.Eq{Columns.Subsystem: subsystem}).ToSql()
	if err != nil {
		logging.Agent.AddSimpleError("Получения Id подсистемы", "Не удалось сформировать sql запрос: "+ err.Error())
		return
	}

	err = repository.pool.
		QueryRow(ctx, query, args...).
		Scan(&subsystem_id)

	return subsystem_id, err
}

func (repository *DBRepository) AddTemplate(ctx context.Context, template *Template) error {

	subsystem_id, err := repository.GetSubsystemId(ctx, template.Subsystem)
	if subsystem_id == 0 {
		return err
	}
	sqr_query := repository.psql.Insert(Tables.Templates)
	if template.GetID() != 0 {
		sqr_query = sqr_query.Columns(Columns.ID, Columns.Name, Columns.Content, Columns.Subsystem, Columns.IsActive, Columns.RenderData, Columns.IsSingle).
		Values(template.ID, template.Name, template.Content, subsystem_id, template.IsActive, template.RenderData, template.IsSingle).Suffix(
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
		Values(template.Name, template.Content, subsystem_id, template.IsActive, template.RenderData, template.IsSingle).Suffix(fmt.Sprintf("RETURNING %s;", Columns.ID))
	}

	query, args, err := sqr_query.ToSql()
	if err != nil {
		logging.Agent.AddSimpleError("Добавление/обновление шаблона", "Не удалось сформировать sql запрос: "+ err.Error())
		return err
	}
	err = repository.pool.QueryRow(ctx, query, args...).Scan(&template.ID)
	return err
}

func (repository *DBRepository) AddSubsystem(ctx context.Context, subsystem_name string, subsystem_path string) error {
	query, args, err := repository.psql.Insert(Tables.Subsystems).Columns(Columns.Subsystem, Columns.RequestPath).
	Values(subsystem_name, subsystem_path).Suffix(fmt.Sprintf("ON CONFLICT (%s) DO NOTHING;", Columns.ID)).ToSql()
	if err != nil {
		logging.Agent.AddSimpleError("Добавление/обновление подсистемы", "Не удалось сформировать sql запрос: "+ err.Error())
		return err
	}

	_, err = repository.pool.Exec(ctx, query, args...)

	return err
}

func (repository *DBRepository) AddTag(ctx context.Context, tag *Tag) error {
	subsystem_id, err := repository.GetSubsystemId(ctx, tag.Subsystem)
	query, args, err := repository.psql.Insert(Tables.Tags).
		Columns(Columns.Name, Columns.Description, Columns.Subsystem, Columns.Alias).
		Values(tag.Name, tag.Description, subsystem_id, tag.Alias).Suffix(
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
		logging.Agent.AddSimpleError("Добавление/обновление тэга", "Не удалось сформировать sql запрос: "+ err.Error())
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
		%s INT,
		%s TEXT NOT NULL,
		%s BOOLEAN DEFAULT TRUE,

		CONSTRAINT fk_tags_subsystems 
			FOREIGN KEY (%s) 
			REFERENCES %s(%s)
		);
	`,
		Tables.Tags,
		Columns.ID,
		Columns.Name,
		Columns.Description,
		Columns.Subsystem,
		Columns.Alias,
		Columns.IsActive,
		Columns.Subsystem,
		Tables.Subsystems, Columns.ID,
	)
	_, err := repository.pool.Exec(ctx, query)
	return err
}

func (repository *DBRepository) createTemplatesTables(ctx context.Context) error {
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		%s SERIAL PRIMARY KEY,
		%s TEXT NOT NULL,
		%s TEXT NOT NULL,
		%s INT,
		%s BOOLEAN DEFAULT TRUE,
		%s JSONB DEFAULT '{}',
		%s BOOLEAN DEFAULT FALSE,

		CONSTRAINT fk_templates_subsystems 
			FOREIGN KEY (%s) 
			REFERENCES %s(%s)
		);
	`,
		Tables.Templates,
		Columns.ID,
		Columns.Name,
		Columns.Content,
		Columns.Subsystem,
		Columns.IsActive,
		Columns.RenderData,
		Columns.IsSingle,
		Columns.Subsystem,
		Tables.Subsystems, Columns.ID,
	)
	_, err := repository.pool.Exec(ctx, query)
	return err
}

func (repository *DBRepository) createSubsystemsTables(ctx context.Context) error {
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		%s SERIAL PRIMARY KEY,
		%s TEXT NOT NULL,
		%s TEXT NOT NULL
		);
	`,
		Tables.Subsystems,
		Columns.ID,
		Columns.Subsystem,
		Columns.RequestPath,
		
	)
	_, err := repository.pool.Exec(ctx, query)
	return err
}