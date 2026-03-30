package handlers

import (
	"encoding/json"
	"former/internal/transformer"
	"net/http"
	"former/internal/db"
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
	raw_data, err := json.Marshal(export_form.Data)
	if err != nil {
		http.Error(writer, "Something went wrong", http.StatusBadRequest)
        return
	}
	export_doc, err:= handler.transformer.RenderTemplate(ctx, export_form.TemplateName, raw_data)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
        return
	}
	json.NewEncoder(writer).Encode(db.Template{Name: export_form.TemplateName, Content: export_doc})
}