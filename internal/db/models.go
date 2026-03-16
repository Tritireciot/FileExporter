package db

type DBModel interface {
	getTable() string
	getName() string
}

type Template struct {
	Name string `json:"name"`
	Content string `json:"content"`
}

func (template *Template) getTable() string {
	return Tables.Templates
}

func (template *Template) getName() string {
	return template.Name
}

type Tag struct {
	Name string `json:"name"`
	Description string `json:"description"`
	Subsystem string `json:"subsystem"`
	Alias string `json:"alias"`
}

func (tag *Tag) getTable() string {
	return Tables.Tags
}

func (tag *Tag) getName() string {
	return tag.Name
}