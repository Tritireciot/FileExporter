package db

type DBModel interface {
	GetTable() string
	GetID() int
	GetName() string
	GetColumns() []any
	Validate() bool
}

type Subsystem string


type Template struct {
	ID         int
	Name       string         `json:"name"`
	Content    string         `json:"content"`
	Subsystem  string         `json:"subsystem"`
	IsActive   bool           `json:"is_active"`
	RenderData map[string]any `json:"render_data"`
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
