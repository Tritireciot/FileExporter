package handlers

import "errors"

type ExportModel struct {
	TemplateId int    `json:"template_id"`
	Data       any    `json:"data"`
	Subsystem  string `json:"subsystem"`
	Format     string `json:"format"`
}

func (model *ExportModel) Validate() (bool, error) {
	is_valid :=  model.TemplateId > 0 && len(model.Subsystem) > 0
	if is_valid {
		return true, nil
	}

	err_message := ""
	if model.TemplateId < 0 {
		err_message += "Reqired Template ID\n"
	}

	if len(model.Subsystem) < 0 {
		err_message += "Reqired Subsystem\n"
	}
	return false, errors.New(err_message)
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
