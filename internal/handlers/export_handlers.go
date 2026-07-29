package handlers

import (
	"PrintServer/internal/transformer"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
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

	var contentType string
	var fileBytes []byte
	var err error

	ctx := request.Context()
	fileName, export_doc, err := handler.transformer.RenderTemplate(ctx, export_form.TemplateId, export_form.Data)
	if err != nil {
		unknownError(writer, err)
		return
	}

	switch export_form.Format {
		case "pdf":
			contentType = "application/pdf"
			fileName += ".pdf"
			fileBytes, err =  handler.transformer.PDFFromTemplate(export_doc)
		case "docx":
			contentType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
			fileName += ".docx"
			fileBytes, err = handler.transformer.DOCXFromTemplate(export_doc)
		default:
			contentType = "text/html"
			fileName += ".html"
			fileBytes = []byte(export_doc)
	}

	if err != nil {
		unknownError(writer, err)
		return
	}

	encodedFileName := url.QueryEscape(fileName)

	writer.Header().Set("Content-Type", contentType)
	writer.Header().Set("Content-Length", fmt.Sprintf("%d", len(fileBytes)))

	writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", encodedFileName))

	writer.Header().Set("X-File-Name", encodedFileName)
	writer.Header().Set("Access-Control-Expose-Headers", "X-File-Name, Content-Disposition")

	writer.Write(fileBytes)
}