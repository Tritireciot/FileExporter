package handlers

import (
	"PrintServer/PrintServer/diagnostic"
	"PrintServer/PrintServer/transformer"
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
		WriteError(
			writer, 
			diagnostic.NewRequestError(
				http.StatusUnprocessableEntity,
				diagnostic.CodeInvalidJSONError,
				err.Error(),
			),
		)
		return
	}
	defer request.Body.Close()

	var contentType string
	var fileBytes []byte
	var err error

	fileName, export_doc, err := handler.transformer.RenderTemplate(export_form.TemplateId, export_form.Subsystem, export_form.Data)
	if err != nil {
		WriteError(writer, err)
		return
	}

	switch export_form.Format {
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
		WriteError(writer, err)
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

func (handler *ExportHandler) PreviewTemplate(writer http.ResponseWriter, request *http.Request) {
	var preview_form PreviewModel
	if err := json.NewDecoder(request.Body).Decode(&preview_form); err != nil {
		WriteError(
			writer, 
			diagnostic.NewRequestError(
				http.StatusUnprocessableEntity,
				diagnostic.CodeInvalidJSONError,
				err.Error(),
			),
		)
		return
	}
	defer request.Body.Close()

	content, name, err := handler.transformer.PreviewTemplate(preview_form.TemplateContent, preview_form.Subsystem, preview_form.TemplateType)

	if err != nil {
		WriteError(writer, err)
		return
	}

	fileBytes := []byte(content)
	encodedFileName := url.QueryEscape(name)

	writer.Header().Set("Content-Type", "text/html")
	writer.Header().Set("Content-Length", fmt.Sprintf("%d", len(fileBytes)))

	writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", encodedFileName))
	writer.Header().Set("X-File-Name", encodedFileName)
	writer.Header().Set("Access-Control-Expose-Headers", "X-File-Name, Content-Disposition")

	writer.Write(fileBytes)

}

func (handler *ExportHandler) ValidateTemplate(writer http.ResponseWriter, request *http.Request) {
	var validation_form ValidateTemplateModel
	if err := json.NewDecoder(request.Body).Decode(&validation_form); err != nil {
		WriteError(
			writer, 
			diagnostic.NewRequestError(
				http.StatusUnprocessableEntity,
				diagnostic.CodeInvalidJSONError,
				err.Error(),
			),
		)
		return
	}
	defer request.Body.Close()

	err := handler.transformer.ValidateTemplate(
		validation_form.TemplateContent,
		validation_form.Subsystem,
		validation_form.TemplateType,
		validation_form.ContentHash,
		validation_form.TemplateData,
	)

	if err != nil {
		WriteError(writer, err)
		return
	}

}
