package handlers

import (
	"encoding/json"
	"errors"
	"former/internal/db"
	"net/http"
)

type TagsHandler struct {
	repo db.DBRepo
}

func NewTagsHandler(repo db.DBRepo) *TagsHandler {
	return &TagsHandler{repo: repo}
}

func (handler *TagsHandler) GetAllTags(writer http.ResponseWriter, request *http.Request) {

	ctx := request.Context()
	tags_list, err := handler.repo.GetAllElements(ctx, &db.Tag{}, "", true)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(tags_list)
}

func (handler *TagsHandler) AddNewTag(writer http.ResponseWriter, request *http.Request) {

	var tag db.Tag
	if err := json.NewDecoder(request.Body).Decode(&tag); err != nil || !tag.Validate() {
		invalidJsonError(writer)
		return
	}

	defer request.Body.Close()
	ctx := request.Context()

	if err := handler.repo.AddTag(ctx, &tag); err != nil {
		failedCreatingError(writer)
		return
	}

	writer.WriteHeader(http.StatusOK)
	json.NewEncoder(writer).Encode(tag)
}

func (handler *TagsHandler) DeleteTag(writer http.ResponseWriter, request *http.Request) {
	tag_id, ok := validateId(request)
	if !ok {
		idRequiredError(writer)
		return
	}

	ctx := request.Context()
	err := handler.repo.DeleteElement(ctx, &db.Tag{ID: tag_id}, db.Columns.ID)
	if err != nil {
		if errors.Is(err, db.ErrElementNotFound) {
			notFoundError(writer)
			return
		}
		unknownError(writer, err)
		return
	}

	writer.WriteHeader(http.StatusOK)
	json.NewEncoder(writer).Encode(IDModel{ID: tag_id})
}
