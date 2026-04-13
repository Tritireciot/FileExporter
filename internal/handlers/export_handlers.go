package handlers

import (
	"encoding/json"
	"former/internal/db"
	"former/internal/transformer"
	"net/http"
)

type ExportHandler struct {
	transformer *transformer.TransformService
}

func NewExportHandler(transformer_ *transformer.TransformService) *ExportHandler {
	return &ExportHandler{transformer: transformer_}
}

func (handler *ExportHandler) TransformTemplate(writer http.ResponseWriter, request *http.Request) {
	var export_form ExportModel
	if err := json.NewDecoder(request.Body).Decode(&export_form); err != nil {
		invalidJsonError(writer)
		return
	}
	defer request.Body.Close()
	ctx := request.Context()
	raw_data, err := json.Marshal(export_form.Data)
	if err != nil {
		invalidJsonError(writer)
		return
	}
	export_doc, err := handler.transformer.RenderTemplate(ctx, export_form.TemplateId, raw_data)
	if err != nil {
		unknownError(writer, err)
		return
	}
	json.NewEncoder(writer).Encode(db.Template{ID: export_form.TemplateId, Content: export_doc})
}
