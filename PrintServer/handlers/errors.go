package handlers

import (
	"net/http"
)

func invalidJsonError(writer http.ResponseWriter) {
	http.Error(writer, "invalid json", http.StatusUnprocessableEntity)
}

func unknownError(writer http.ResponseWriter, err error) {
	http.Error(writer, err.Error(), http.StatusInternalServerError)
}
