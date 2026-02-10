package handlers

import (
	"encoding/json"
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
		http.Error(writer, "invalid json", http.StatusBadRequest)
        return
	}
	defer request.Body.Close()
	ctx := request.Context()
	export_doc, err:= handler.transformer.RenderTemplate(ctx, export_form.TemplateName, export_form.Data)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
        return
	}
	writer.Write([]byte(export_doc))
}