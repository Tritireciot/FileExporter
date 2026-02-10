package handlers

type ExportModel struct {
	TemplateName string `json:"template_name"`
	Data map[string]any `json:"data"`
}