package handlers

import (
	"net/http"
)

func idRequiredError(writer http.ResponseWriter) {
	http.Error(writer, "id is required", http.StatusUnprocessableEntity)
}

func invalidJsonError(writer http.ResponseWriter) {
	http.Error(writer, "invalid json", http.StatusUnprocessableEntity)
}

func unknownError(writer http.ResponseWriter, err error) {
	http.Error(writer, err.Error(), http.StatusInternalServerError)
}

func failedCreatingError(writer http.ResponseWriter) {
	http.Error(writer, "failed to create", http.StatusInternalServerError)
}

func notFoundError(writer http.ResponseWriter) {
	http.Error(writer, "not found", http.StatusNotFound)
}
