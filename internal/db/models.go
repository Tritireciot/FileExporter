package db

type Template struct {
	Name string `json:"name"`
	Content string `json:"content"`
}

type Tag struct {
	Name string `json:"name"`
	Description string `json:"description"`
	Subsystem string `json:"subsystem"`
}