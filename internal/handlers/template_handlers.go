package handlers

import (
	"PrintServer/internal/db"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type TemplateHandler struct {
	repo db.DBRepo
}

func NewTemplateHandler(repo db.DBRepo) *TemplateHandler {
	return &TemplateHandler{repo: repo}
}

func (handler *TemplateHandler) GetAllTemplates(writer http.ResponseWriter, request *http.Request) {
	subsystem_name := request.URL.Query().Get("subsystem")
	active, _ := validateFlag(request, "active")
	ctx := request.Context()
	template_list, err := handler.repo.GetAllElements(ctx, &db.Template{}, subsystem_name, active)
	if err != nil {
		unknownError(writer, err)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(template_list)
}

func (handler *TemplateHandler) GetTemplate(writer http.ResponseWriter, request *http.Request) {
	template_id, ok := validateId(request)
	if !ok {
		idRequiredError(writer)
		return
	}
	ctx := request.Context()
	template := db.Template{ID: template_id}
	err := handler.repo.GetElement(ctx, &template, db.Columns.ID)
	if err != nil {
		if errors.Is(err, db.ErrElementNotFound) {
			notFoundError(writer)
			return
		}
		unknownError(writer, err)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(template)
}

func (handler *TemplateHandler) AddNewTemplate(writer http.ResponseWriter, request *http.Request) {

	var template db.Template
	if err := json.NewDecoder(request.Body).Decode(&template); err != nil || !template.Validate() {
		invalidJsonError(writer)
		return
	}

	defer request.Body.Close()
	template.Subsystem = strings.ToLower(template.Subsystem)
	ctx := request.Context()

	if err := handler.repo.AddTemplate(ctx, &template); err != nil {
		failedCreatingError(writer)
		return
	}

	writer.WriteHeader(http.StatusOK)
	json.NewEncoder(writer).Encode(template)
}

func (handler *TemplateHandler) DeleteTemplate(writer http.ResponseWriter, request *http.Request) {
	template_id, ok := validateId(request)
	if !ok {
		idRequiredError(writer)
		return
	}

	ctx := request.Context()

	err := handler.repo.DeleteElement(ctx, &db.Template{ID: template_id}, db.Columns.ID)
	if err != nil {
		if errors.Is(err, db.ErrElementNotFound) {
			notFoundError(writer)
			return
		}
		unknownError(writer, err)
		return
	}

	writer.WriteHeader(http.StatusOK)
	json.NewEncoder(writer).Encode(IDModel{ID: template_id})
}
