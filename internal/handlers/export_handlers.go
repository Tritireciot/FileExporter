package handlers

import (
	"PrintServer/internal/db"
	"PrintServer/internal/transformer"
	"encoding/json"
	"net/http"
)

type ExportHandler struct {
	transformer transformer.Transformer
}

func NewExportHandler(transformer_ transformer.Transformer) *ExportHandler {
	return &ExportHandler{transformer: transformer_}
}

func (handler *ExportHandler) TransformTemplate(writer http.ResponseWriter, request *http.Request) {
	var export_form ExportModel
	if err := json.NewDecoder(request.Body).Decode(&export_form); err != nil || !export_form.Validate() {
		invalidJsonError(writer)
		return
	}
	defer request.Body.Close()
	ctx := request.Context()
	title, export_doc, err := handler.transformer.RenderTemplate(ctx, export_form.TemplateId, export_form.Data, export_form.Subsystem)
	if err != nil {
		unknownError(writer, err)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(db.Template{ID: export_form.TemplateId, Name: title, Content: export_doc, Subsystem: export_form.Subsystem})
}
