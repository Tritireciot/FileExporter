package handlers

import (
	"net/http"
	"strconv"
)

func validateId(request *http.Request) (int, bool) {
	id_, err := strconv.Atoi(request.URL.Query().Get("id"))
	if (id_ < 0 && err == nil) || err != nil {
		return id_, false
	}
	return id_, true
}