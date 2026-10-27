package handlers

type ExportModel struct {
	TemplateId int    `json:"template_id"`
	Data       any    `json:"data"`
	Subsystem  string `json:"subsystem"`
	Format     string `json:"format"`
}

func (model *ExportModel) Validate() bool {
	return model.TemplateId > 0 && len(model.Subsystem) > 0
}

type PreviewModel struct {
	Subsystem       string `json:"subsystem"`
	TemplateType    string `json:"template_type"`
	TemplateContent string `json:"template"`
}

type ValidateTemplateModel struct {
	Subsystem       string         `json:"subsystem"`
	TemplateType    string         `json:"template_type"`
	TemplateContent string         `json:"template"`
	ContentHash     string         `json:"content_hash"`
	TemplateData    map[string]any `json:"data"`
}
