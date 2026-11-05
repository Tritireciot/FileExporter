package handlers

import (
	"PrintServer/PrintServer/diagnostic"
	"PrintServer/PrintServer/transformer"
	logging "PrintServer/agent"
	"encoding/json"
	"errors"
	"net"
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
			Code:    reqErr.Code,
			Message: reqErr.Message,
			Details: ErrorDetails{
				Stage:     diagnostic.StageRequest,
				Retryable: reqErr.Status >= http.StatusInternalServerError,
			},
		}
	}

	var templateTypeNotFound *transformer.TemplateTypeNotFoundError
	if errors.As(err, &templateTypeNotFound) {
		return lookupError(http.StatusNotFound, diagnostic.CodeTemplateTypeNotFound, "Тип шаблона не найден", false)
	}
	var templateNotFound *transformer.TemplateNotFoundError
	if errors.As(err, &templateNotFound) {
		return lookupError(http.StatusNotFound, diagnostic.CodeTemplateNotFound, "Шаблон не найден", false)
	}
	var previewFetch *transformer.PreviewDataFetchError
	if errors.As(err, &previewFetch) {
		var netErr net.Error
		if errors.As(previewFetch.Cause, &netErr) {
			return lookupError(http.StatusBadGateway, diagnostic.CodeConfigServerUnavailable, "ConfigAppServer недоступен", true)
		}
		return lookupError(http.StatusBadGateway, diagnostic.CodePreviewDataFetchFailed, "Не удалось получить данные предпросмотра", true)
	}
	var templateFetch *transformer.TemplateFetchError
	if errors.As(err, &templateFetch) {
		var statusErr *transformer.UpstreamStatusError
		if errors.As(templateFetch.Cause, &statusErr) {
			switch statusErr.StatusCode {
			case http.StatusNotFound:
				return lookupError(http.StatusNotFound, diagnostic.CodeTemplateNotFound, "Шаблон не найден", false)
			case http.StatusConflict:
				return lookupError(http.StatusConflict, diagnostic.CodeTemplateNotActive, "Шаблон не опубликован", false)
			}
		}
		var netErr net.Error
		if errors.As(templateFetch.Cause, &netErr) {
			return lookupError(http.StatusBadGateway, diagnostic.CodeConfigServerUnavailable, "ConfigAppServer недоступен", true)
		}
		return lookupError(http.StatusBadGateway, diagnostic.CodeTemplateFetchFailed, "Не удалось получить шаблон", true)
	}

	return mapInternalError()
}

func lookupError(status int, code diagnostic.Code, message string, retryable bool) (int, ErrorResponse) {
	return status, ErrorResponse{
		Code: code, Message: message,
		Details: ErrorDetails{Stage: diagnostic.StageTemplateLookup, Retryable: retryable},
	}
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
	diagnostic.CodeUnclosedSpecialTag: {
		Status: http.StatusUnprocessableEntity, Message: "Специальная конструкция шаблона не закрыта", Field: string(ErrorFieldTemplate),
	},
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
	diagnostic.CodeElseWithoutIf: {
		Status: http.StatusUnprocessableEntity, Message: "Else используется без открытого If", Field: string(ErrorFieldTemplate),
	},
	diagnostic.CodeElseIfWithoutIf: {
		Status: http.StatusUnprocessableEntity, Message: "ElseIf используется без открытого If", Field: string(ErrorFieldTemplate),
	},
	diagnostic.CodeDuplicateElse: {
		Status: http.StatusUnprocessableEntity, Message: "В блоке If указано несколько Else", Field: string(ErrorFieldTemplate),
	},
	diagnostic.CodeElseIfAfterElse: {
		Status: http.StatusUnprocessableEntity, Message: "ElseIf не может следовать после Else", Field: string(ErrorFieldTemplate),
	},
	diagnostic.CodeInvalidRepeat: {
		Status: http.StatusUnprocessableEntity, Message: "Некорректная конструкция Repeat", Field: string(ErrorFieldTemplate),
	},
	diagnostic.CodeInvalidWith: {
		Status: http.StatusUnprocessableEntity, Message: "Некорректная конструкция With", Field: string(ErrorFieldTemplate),
	},
	diagnostic.CodeInvalidPath: {
		Status: http.StatusUnprocessableEntity, Message: "Некорректный путь к полю", Field: string(ErrorFieldTemplate),
	},
	diagnostic.CodeEmptyIndex: {
		Status: http.StatusUnprocessableEntity, Message: "Индекс массива не указан", Field: string(ErrorFieldTemplate),
	},
	diagnostic.CodeUnclosedIndex: {
		Status: http.StatusUnprocessableEntity, Message: "Индекс массива не закрыт", Field: string(ErrorFieldTemplate),
	},
	diagnostic.CodeInvalidCondition: {
		Status: http.StatusUnprocessableEntity, Message: "Некорректное условие", Field: string(ErrorFieldTemplate),
	},
	diagnostic.CodeUnclosedCondition: {
		Status: http.StatusUnprocessableEntity, Message: "Условие не закрыто", Field: string(ErrorFieldTemplate),
	},
	diagnostic.CodeInvalidField: {
		Status: http.StatusUnprocessableEntity, Message: "Некорректное поле шаблона", Field: string(ErrorFieldTemplate),
	},
	diagnostic.CodeInvalidFieldPipeline: {
		Status: http.StatusUnprocessableEntity, Message: "Некорректный конвейер обработки поля", Field: string(ErrorFieldTemplate),
	},
	diagnostic.CodeExpectedOperand: {
		Status: http.StatusUnprocessableEntity, Message: "В условии ожидается операнд", Field: string(ErrorFieldTemplate),
	},
	diagnostic.CodeUnmatchedParenthesis: {
		Status: http.StatusUnprocessableEntity, Message: "Скобки в условии не согласованы", Field: string(ErrorFieldTemplate),
	},
	diagnostic.CodeUnknownSpecialTag: {
		Status: http.StatusUnprocessableEntity, Message: "Неизвестная специальная конструкция шаблона", Field: string(ErrorFieldTemplate),
	},
	diagnostic.CodeEmptyField: {
		Status: http.StatusUnprocessableEntity, Message: "Имя поля шаблона не указано", Field: string(ErrorFieldTemplate),
	},
	diagnostic.CodeGoTemplateParse: {
		Status: http.StatusUnprocessableEntity, Message: "Не удалось разобрать шаблон после преобразования", Field: string(ErrorFieldTemplate),
	},
	diagnostic.CodeTemplateFieldNotFound: {
		Status:    http.StatusUnprocessableEntity,
		Message:   "В переданных данных не найдено указанное поле",
		Field:     string(ErrorFieldData),
		Retryable: false,
	},
	diagnostic.CodeTemplateFunctionError: {
		Status: http.StatusUnprocessableEntity, Message: "Ошибка выполнения функции шаблона", Field: string(ErrorFieldTemplate),
	},
	diagnostic.CodeTemplateSyntaxError: {
		Status: http.StatusUnprocessableEntity, Message: "Синтаксическая ошибка шаблона", Field: string(ErrorFieldTemplate),
	},
	diagnostic.CodeInvalidFormatter: {
		Status: http.StatusUnprocessableEntity, Message: "Некорректный форматтер шаблона", Field: string(ErrorFieldTemplate),
	},
	diagnostic.CodeTemplateExecutionFailed: {
		Status:    http.StatusUnprocessableEntity,
		Message:   "Не удалось выполнить шаблон",
		Retryable: false,
	},
	diagnostic.CodeInvalidJSONError: {
		Status: http.StatusBadRequest, Message: "Некорректный JSON в запросе", Field: string(ErrorFieldData),
	},
	diagnostic.CodeInternalError: {
		Status: http.StatusInternalServerError, Message: "Внутренняя ошибка сервера",
	},
	diagnostic.CodeInvalidNumber: {
		Status: http.StatusUnprocessableEntity, Message: "Некорректное число", Field: string(ErrorFieldTemplate),
	},
	diagnostic.CodeUnclosedString: {
		Status: http.StatusUnprocessableEntity, Message: "Строковый литерал некорректно записан", Field: string(ErrorFieldTemplate),
	},
}

func WriteError(
	writer http.ResponseWriter,
	err error,
) {
	status, response := MapError(err)

	errMessage := response.Message
	if err != nil {
		errMessage = err.Error()
	}

	if status >= http.StatusInternalServerError {
		logging.Agent.AddSimpleError(
			"Обработка HTTP-запроса",
			errMessage,
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
