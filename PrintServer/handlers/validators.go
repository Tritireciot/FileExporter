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

func validateFlag(request *http.Request, flagName string) (bool, bool) {
	flag := request.URL.Query().Get(flagName)
	if flag == "" {
		return false, false
	}
	bool_flag, err := strconv.ParseBool(flag)
	if err != nil {
		return false, false
	}
	return bool_flag, true
}
