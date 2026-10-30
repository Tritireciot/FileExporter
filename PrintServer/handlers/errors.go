package handlers

import (
	logging "AutoplayX/agent"
	"PrintServer/PrintServer/diagnostic"
	"encoding/json"
	"errors"
	"net/http"
)

type ErrorResponse struct {
	Code    diagnostic.Code     `json:"code"`
	Message string              `json:"message"`
	Fields  map[string][]string `json:"fields,omitempty"`
	Details ErrorDetails        `json:"details,omitempty"`
}

type ErrorDetails struct {
	Stage     diagnostic.Stage `json:"stage"`
	Hint      string           `json:"hint,omitempty"`
	Span      *diagnostic.Span `json:"span,omitempty"`
	Retryable bool             `json:"retryable"`
}

func MapError(err error) (int, ErrorResponse) {
	if err == nil {
		return mapInternalError()
	}

	var templateErr *diagnostic.TemplateError

	if errors.As(err, &templateErr) {
		return mapTemplateError(templateErr)
	}

	var reqErr *diagnostic.RequestError

	if errors.As(err, &reqErr) {
		return reqErr.Status, ErrorResponse{
			Code: reqErr.Code,
			Message: reqErr.Message,
			Details: ErrorDetails{
				Stage: diagnostic.StageRequest,
				Retryable: true,
			},
		}
	}

	return mapInternalError()
}

type ErrorField string

const (
	ErrorFieldTemplate ErrorField = "template"
	ErrorFieldData     ErrorField = "data"
)

type errorPolicy struct {
	Status    int
	Message   string
	Field     string
	Retryable bool
}

var errorPolicies = map[diagnostic.Code]errorPolicy{
	diagnostic.CodeUnclosedBlock: {
		Status:    http.StatusUnprocessableEntity,
		Message:   "Блок шаблона не закрыт",
		Field:     string(ErrorFieldTemplate),
		Retryable: false,
	},
	diagnostic.CodeUnexpectedEnd: {
		Status:    http.StatusUnprocessableEntity,
		Message:   "В шаблоне обнаружено лишнее окончание блока",
		Field:     string(ErrorFieldTemplate),
		Retryable: false,
	},
	diagnostic.CodeInvalidIndex: {
		Status:    http.StatusUnprocessableEntity,
		Message:   "В шаблоне указан некорректный индекс",
		Field:     string(ErrorFieldTemplate),
		Retryable: false,
	},
	diagnostic.CodeTemplateFieldNotFound: {
		Status:    http.StatusUnprocessableEntity,
		Message:   "В переданных данных не найдено указанное поле",
		Field:     string(ErrorFieldData),
		Retryable: false,
	},
	diagnostic.CodeTemplateExecutionFailed: {
		Status:    http.StatusUnprocessableEntity,
		Message:   "Не удалось выполнить шаблон",
		Retryable: false,
	},
}

func WriteError(
	writer http.ResponseWriter,
	err error,
) {
	status, response := MapError(err)

	if status >= http.StatusInternalServerError {
		logging.Agent.AddSimpleError(
			"Обработка HTTP-запроса",
			err.Error(),
		)
	}

	writer.Header().Set(
		"Content-Type",
		"application/json; charset=utf-8",
	)
	writer.WriteHeader(status)

	encodeErr := json.NewEncoder(writer).Encode(response)
	if encodeErr != nil {
		logging.Agent.AddSimpleError(
			"Формирование HTTP-ответа",
			encodeErr.Error(),
		)
	}
}
