package handlers

import (
	"PrintServer/PrintServer/diagnostic"
	"net/http"
	"regexp"
)

var routeNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_-]*$`)
var contentHashPattern = regexp.MustCompile(`^[a-f0-9]{64}$`)

func validRouteName(value string) bool {
	return len(value) <= 64 && routeNamePattern.MatchString(value)
}

type ExportModel struct {
	TemplateId int    `json:"template_id"`
	Data       any    `json:"data"`
	Subsystem  string `json:"subsystem"`
	Format     string `json:"format"`
}

func (model *ExportModel) Validate() (bool, error) {
	if model.TemplateId <= 0 {
		return false, invalidRequest("template_id должен быть положительным числом")
	}
	if !validRouteName(model.Subsystem) {
		return false, invalidRequest("Некорректное имя подсистемы")
	}
	if _, ok := model.Data.(map[string]any); !ok {
		return false, invalidRequest("data должен быть объектом JSON")
	}
	if model.Format != "" && model.Format != "html" && model.Format != "docx" {
		return false, diagnostic.NewRequestError(http.StatusUnprocessableEntity, diagnostic.CodeUnsupportedFormat, "Неподдерживаемый формат документа")
	}
	return true, nil
}

type PreviewModel struct {
	Subsystem       string `json:"subsystem"`
	TemplateType    string `json:"template_type"`
	TemplateContent string `json:"template"`
}

func (model *PreviewModel) Validate() error {
	if !validRouteName(model.Subsystem) || !validRouteName(model.TemplateType) {
		return invalidRequest("Некорректная подсистема или тип шаблона")
	}
	return nil
}

type ValidateTemplateModel struct {
	TemplateID      int64          `json:"template_id"`
	Subsystem       string         `json:"subsystem"`
	TemplateType    string         `json:"template_type"`
	TemplateContent string         `json:"template"`
	ContentHash     string         `json:"content_hash"`
	TemplateData    map[string]any `json:"data"`
}

func (model *ValidateTemplateModel) Validate() error {
	if model.TemplateID <= 0 {
		return invalidRequest("template_id должен быть положительным числом")
	}
	if !validRouteName(model.Subsystem) || !validRouteName(model.TemplateType) {
		return invalidRequest("Некорректная подсистема или тип шаблона")
	}
	if !contentHashPattern.MatchString(model.ContentHash) {
		return invalidRequest("content_hash должен содержать 64 строчных шестнадцатеричных символа")
	}
	if model.TemplateData == nil {
		return invalidRequest("data должен быть объектом JSON")
	}
	return nil
}
