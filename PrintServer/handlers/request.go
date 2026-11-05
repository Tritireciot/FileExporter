package handlers

import (
	"PrintServer/PrintServer/diagnostic"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
)

func decodeRequest(request *http.Request, target any, required, optional []string) error {
	mediaType, _, err := mime.ParseMediaType(request.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		return diagnostic.NewRequestError(http.StatusUnsupportedMediaType, diagnostic.CodeUnsupportedMediaType, "Ожидается Content-Type: application/json")
	}

	body, err := io.ReadAll(request.Body)
	if err != nil {
		return diagnostic.NewRequestError(http.StatusBadRequest, diagnostic.CodeInvalidJSONError, err.Error())
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(body, &fields); err != nil {
		return diagnostic.NewRequestError(http.StatusBadRequest, diagnostic.CodeInvalidJSONError, err.Error())
	}
	if fields == nil {
		return invalidRequest("Ожидается объект JSON")
	}

	allowed := make(map[string]bool, len(required)+len(optional))
	for _, field := range required {
		allowed[field] = true
		value, ok := fields[field]
		if !ok || bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return invalidRequest(fmt.Sprintf("Не указано обязательное поле %q", field))
		}
	}
	for _, field := range optional {
		allowed[field] = true
	}
	for field, value := range fields {
		if !allowed[field] {
			return invalidRequest(fmt.Sprintf("Неизвестное поле %q", field))
		}
		if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return invalidRequest(fmt.Sprintf("Поле %q не может быть null", field))
		}
	}
	if err := json.Unmarshal(body, target); err != nil {
		return invalidRequest(err.Error())
	}
	return nil
}

func invalidRequest(message string) error {
	return diagnostic.NewRequestError(http.StatusUnprocessableEntity, diagnostic.CodeInvalidRequest, message)
}
