package handlers

import (
	"encoding/json"
	"errors"
	"former/internal/db"
	"net/http"
)

type TagsHandler struct {
	repo *db.DBRepository
}

func NewTagsHandler(repo *db.DBRepository) *TagsHandler {
	return &TagsHandler{repo: repo}
}

func (handler *TagsHandler) GetAllTags(writer http.ResponseWriter, request *http.Request){
	
	ctx := request.Context()
	tags_list, err := handler.repo.GetAllElements(ctx, db.Tables.Tags)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	

	writer.Header().Set("Content-Type", "application/json")
    json.NewEncoder(writer).Encode(tags_list)
}

func (handler *TagsHandler) AddNewTag(writer http.ResponseWriter, request *http.Request){

	var tag db.Tag
	if err := json.NewDecoder(request.Body).Decode(&tag); err != nil {
		http.Error(writer, "invalid json", http.StatusBadRequest)
        return
	}

	defer request.Body.Close()
	ctx := request.Context()
	
	if err := handler.repo.AddTag(ctx, &tag); err != nil {
		http.Error(writer, "failed to create tag", http.StatusInternalServerError)
        return
	}

	writer.WriteHeader(http.StatusCreated)
}

func (handler *TagsHandler) DeleteTag(writer http.ResponseWriter, request *http.Request) {
	tag_name := request.URL.Query().Get("name")
	if tag_name == "" {
		http.Error(writer, "name is required", http.StatusBadRequest)
		return
	}

	ctx := request.Context()

	err := handler.repo.DeleteElement(ctx, tag_name, db.Tables.Tags)
	if err != nil {
		if errors.Is(err, db.ErrElementNotFound) {
			http.Error(writer, "tag not found", http.StatusNotFound)
			return
		}
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}