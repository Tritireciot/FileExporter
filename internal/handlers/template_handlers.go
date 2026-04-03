package handlers

import (
	"encoding/json"
	"errors"
	"former/internal/db"
	"net/http"
	"strconv"
)

type TemplateHandler struct {
	repo *db.DBRepository
}

func NewTemplateHandler(repo *db.DBRepository) *TemplateHandler {
	return &TemplateHandler{repo: repo}
}

func (handler *TemplateHandler) GetAllTemplates(writer http.ResponseWriter, request *http.Request) {

	ctx := request.Context()
	template_list, err := handler.repo.GetAllElements(ctx, &db.Template{})
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(template_list)
}

func (handler *TemplateHandler) GetTemplate(writer http.ResponseWriter, request *http.Request) {
	template_id, err := strconv.Atoi(request.URL.Query().Get("id"))
	name := request.URL.Query().Get("name")
	if (name == "" && (template_id < 0 && err == nil)) || (name == "" && err != nil) {
		http.Error(writer, "name or id is required", http.StatusBadRequest)
		return
	}
	ctx := request.Context()
	column := db.Columns.ID
	if err != nil || template_id < 0 {
		column = db.Columns.Name
	}
	
	template := db.Template{ID: template_id, Name: name}
	err = handler.repo.GetElement(ctx, &template, column)
	if err != nil {
		if errors.Is(err, db.ErrTemplateNotFound) {
			http.Error(writer, "template not found", http.StatusNotFound)
			return
		}
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(writer).Encode(template)
}

func (handler *TemplateHandler) AddNewTemplate(writer http.ResponseWriter, request *http.Request) {

	var template db.Template
	if err := json.NewDecoder(request.Body).Decode(&template); err != nil {
		http.Error(writer, "invalid json", http.StatusBadRequest)
		return
	}

	defer request.Body.Close()
	ctx := request.Context()

	if err := handler.repo.AddTemplate(ctx, &template); err != nil {
		http.Error(writer, "failed to create template", http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusCreated)
}

func (handler *TemplateHandler) DeleteTemplate(writer http.ResponseWriter, request *http.Request) {
	template_name := request.URL.Query().Get("name")
	if template_name == "" {
		http.Error(writer, "name is required", http.StatusBadRequest)
		return
	}

	ctx := request.Context()

	err := handler.repo.DeleteElement(ctx, &db.Template{Name: template_name})
	if err != nil {
		if errors.Is(err, db.ErrElementNotFound) {
			http.Error(writer, "template not found", http.StatusNotFound)
			return
		}
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}
