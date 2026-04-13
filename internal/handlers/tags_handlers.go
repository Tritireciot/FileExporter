package handlers

import (
	"encoding/json"
	"errors"
	"former/internal/db"
	"net/http"
	"strconv"
)

type TagsHandler struct {
	repo *db.DBRepository
}

func NewTagsHandler(repo *db.DBRepository) *TagsHandler {
	return &TagsHandler{repo: repo}
}

func (handler *TagsHandler) GetAllTags(writer http.ResponseWriter, request *http.Request) {

	ctx := request.Context()
	tags_list, err := handler.repo.GetAllElements(ctx, &db.Tag{})
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(tags_list)
}

func (handler *TagsHandler) AddNewTag(writer http.ResponseWriter, request *http.Request) {

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
	if err := handler.repo.GetElement(ctx, &tag, db.Columns.Name); err != nil {
		http.Error(writer, "failed to create tag", http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusCreated)
	json.NewEncoder(writer).Encode(tag)
}

func (handler *TagsHandler) DeleteTag(writer http.ResponseWriter, request *http.Request) {
	tag_id, err := strconv.Atoi(request.URL.Query().Get("id"))
	if (tag_id < 0 && err == nil) || err != nil {
		http.Error(writer, "id is required", http.StatusBadRequest)
		return
	}

	ctx := request.Context()
	err = handler.repo.DeleteElement(ctx, &db.Tag{ID: tag_id}, db.Columns.ID)
	if err != nil {
		if errors.Is(err, db.ErrElementNotFound) {
			http.Error(writer, "tag not found", http.StatusNotFound)
			return
		}
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusOK)
	json.NewEncoder(writer).Encode(map[string]int{"ID": tag_id})
}
