package handlers

type ExportModel struct {
	TemplateId int            `json:"template_id"`
	Data       map[string]any `json:"data"`
}


type IDModel struct {
	ID int `json:"ID"`
}