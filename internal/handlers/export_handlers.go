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
	rundown_title := ""
	if rundown, ok := export_form.Data["rundown"].(map[string]any); ok {
		rundown_title, _ = rundown["title"].(string)
	}
	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(db.Template{ID: export_form.TemplateId, Name: rundown_title, Content: export_doc})
}
