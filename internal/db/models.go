package db

type DBModel interface {
	GetTable() string
	GetID() int
	GetName() string
	GetColumns() []any
	Validate() bool
}

type Template struct {
	ID      int
	Name    string `json:"name"`
	Content string `json:"content"`
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
		&template.ID, &template.Name, &template.Content,
	}
}
func (template *Template) Validate() bool{
	return template.Name != "" && template.Content != ""
}

type Tag struct {
	ID          int
	Name        string `json:"name"`
	Description string `json:"description"`
	Subsystem   string `json:"subsystem"`
	Alias       string `json:"alias"`
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
		&tag.ID, &tag.Name, &tag.Description, &tag.Subsystem, &tag.Alias,
	}
}

func (tag *Tag) Validate() bool{
	return tag.Name != "" && tag.Description != "" && 
	tag.Subsystem != "" && tag.Alias != ""
}
