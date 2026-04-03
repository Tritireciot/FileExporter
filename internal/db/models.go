package db



type DBModel interface {
	getTable() string
	getID() int
	getName() string
	getColumns() []any
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

func (template *Template) getID() int {
	return template.ID
}


func (template *Template) getColumns() []any {
	return []any{
		&template.ID,  &template.Name, &template.Content,
	}
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

func (tag *Tag) getID() int {
	return tag.ID
}


func (tag *Tag) getColumns() []any {
	return []any{
		&tag.ID,  &tag.Name, &tag.Description, &tag.Subsystem, &tag.Alias,
	}
}