package db

var Tables = struct {
	Templates string
	Tags      string
}{
	Templates: "templates",
	Tags:      "tags",
}

var Columns = struct {
	ID          string
	Name        string
	Content     string
	Subsystem   string
	Description string
	Alias       string
}{
	ID:          "id",
	Name:        "name",
	Content:     "content",
	Subsystem:   "subsystem",
	Description: "description",
	Alias:       "alias",
}
