package db

type ColumnsType interface {
	*int, *string, *string, *string, *string | *int, *string, *string // TODO тут исправить как надо с интерфейсами
}

type DBModel interface {
	getTable() string
	getName() string
	getColumns() 
}

type Template struct {
	ID int
	Name string `json:"name"`
	Content string `json:"content"`
}

func (template *Template) getTable() string {
	return Tables.Templates
}

func (template *Template) getName() string {
	return template.Name
}

func (template *Template) getColumns() (*int, *string, *string) {
	return &template.ID,  &template.Name, &template.Content
}

type Tag struct {
	ID int
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

func (tag *Tag) getColumns() (*int, *string, *string, *string, *string) {
	return &tag.ID,  &tag.Name, &tag.Description, &tag.Subsystem, &tag.Alias
}