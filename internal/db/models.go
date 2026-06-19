package db

import (
	"fmt"
)

type DBModel interface {
	GetTable() string
	GetID() int
	GetName() string
	GetColumns() []any
	Validate() bool
	Columns() []string
	GetValueByColumn(column string) any
}

type Subsystem struct {
	ID         int
	Subsystem  string
	RequestPath string
}

func (subsystem *Subsystem) GetValueByColumn(column string) any {
	switch column {
		case Columns.ID:
			return subsystem.ID
		case Columns.Subsystem:
			return subsystem.Subsystem
		case Columns.RequestPath:
			return subsystem.RequestPath
	}
	return nil
}

func (subsystem *Subsystem) GetTable() string {
	return Tables.Subsystems
}

func (subsystem *Subsystem) GetName() string {
	return subsystem.Subsystem
}

func (subsystem *Subsystem) GetID() int {
	return subsystem.ID
}

func (subsystem *Subsystem) GetColumns() []any {
	return []any{
		&subsystem.ID, &subsystem.Subsystem, &subsystem.RequestPath, 
	}
}
func (subsystem *Subsystem) Validate() bool {
	return subsystem.Subsystem != "" && subsystem.RequestPath != ""
}

func (subsystem *Subsystem) Columns() []string {
	return []string{
		fmt.Sprintf("%s.%s", Tables.Subsystems, Columns.ID),
		fmt.Sprintf("%s.%s", Tables.Subsystems, Columns.Subsystem),
		fmt.Sprintf("%s.%s", Tables.Subsystems, Columns.RequestPath),
	}
}

type Template struct {
	ID         int
	Name       string         `json:"name"`
	Content    string         `json:"content"`
	Subsystem  string         `json:"subsystem"`
	IsActive   bool           `json:"is_active"`
	RenderData map[string]bool `json:"render_data"`
	IsSingle   bool           `json:"is_single"`
}

func (template *Template) GetTable() string {
	return Tables.Templates
}

func (template *Template) GetName() string {
	return template.Name
}

func (template *Template) GetID() int {
	return template.ID
}

func (template *Template) GetColumns() []any {
	return []any{
		&template.ID, &template.Name, &template.Content, &template.Subsystem,
		&template.IsActive, &template.RenderData, &template.IsSingle,
	}
}
func (template *Template) Validate() bool {
	return template.Name != "" && template.Content != ""
}

func (template *Template) Columns() []string {
	return []string{
		fmt.Sprintf("%s.%s", Tables.Templates, Columns.ID),
		fmt.Sprintf("%s.%s", Tables.Templates, Columns.Name),
		fmt.Sprintf("%s.%s", Tables.Templates, Columns.Content),
		fmt.Sprintf("%s.%s", Tables.Subsystems, Columns.Subsystem),
		fmt.Sprintf("%s.%s", Tables.Templates, Columns.IsActive),
		fmt.Sprintf("%s.%s", Tables.Templates, Columns.RenderData),
		fmt.Sprintf("%s.%s", Tables.Templates, Columns.IsSingle),
	}
}

func (template *Template) GetValueByColumn(column string) any {
	switch column {
		case Columns.ID:
			return template.ID
		case Columns.Name:
			return template.Name
		case Columns.Subsystem:
			return template.Subsystem
		case Columns.IsActive:
			return template.IsActive
		case Columns.IsSingle:
			return template.IsSingle
	}
	return nil
}

type Tag struct {
	ID          int
	Name        string `json:"name"`
	Description string `json:"description"`
	Subsystem   string `json:"subsystem"`
	Alias       string `json:"alias"`
	IsActive    bool   `json:"is_active"`
}

func (tag *Tag) GetTable() string {
	return Tables.Tags
}

func (tag *Tag) GetName() string {
	return tag.Name
}

func (tag *Tag) GetID() int {
	return tag.ID
}

func (tag *Tag) GetColumns() []any {
	return []any{
		&tag.ID, &tag.Name, &tag.Description, &tag.Subsystem, &tag.Alias, &tag.IsActive,
	}
}

func (tag *Tag) Validate() bool {
	return tag.Name != "" && tag.Description != "" &&
		tag.Subsystem != "" && tag.Alias != ""
}

func (tag *Tag) Columns() []string {
	return []string{
		fmt.Sprintf("%s.%s", Tables.Tags, Columns.ID),
		fmt.Sprintf("%s.%s", Tables.Tags, Columns.Name),
		fmt.Sprintf("%s.%s", Tables.Tags, Columns.Description),
		fmt.Sprintf("%s.%s", Tables.Subsystems, Columns.Subsystem),
		fmt.Sprintf("%s.%s", Tables.Tags, Columns.Alias),
		fmt.Sprintf("%s.%s", Tables.Tags, Columns.IsActive),
	}
}

func (tag *Tag) GetValueByColumn(column string) any {
	switch column {
		case Columns.ID:
			return tag.ID
		case Columns.Name:
			return tag.Name
		case Columns.Subsystem:
			return tag.Subsystem
		case Columns.Alias:
			return tag.Alias
		case Columns.IsActive:
			return tag.IsActive
	}
	return nil
}