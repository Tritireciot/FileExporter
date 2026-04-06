package handlers

type ExportModel struct {
	TemplateId int `json:"template_id"`
	Data map[string]any `json:"data"`
}