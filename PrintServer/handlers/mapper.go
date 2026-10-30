package handlers

import (
	"PrintServer/PrintServer/diagnostic"
	"net/http"
)

func mapTemplateError(
	templateErr *diagnostic.TemplateError,
) (int, ErrorResponse) {
	if templateErr == nil ||
		len(templateErr.Diagnostics) == 0 {
		return mapInternalError()
	}

	item := templateErr.Diagnostics[0]

	policy, exists := errorPolicies[item.Code]
	if !exists {
		return mapInternalError()
	}

	response := ErrorResponse{
		Code:    item.Code,
		Message: policy.Message,
		Details: ErrorDetails{
			Stage:     item.Stage,
			Hint:      item.Hint,
			Span:      item.Span,
			Retryable: policy.Retryable,
		},
	}

	if policy.Field != "" {
		response.Fields = map[string][]string{
			string(policy.Field): {
				item.Message,
			},
		}
	}

	return policy.Status, response
}

func mapInternalError() (int, ErrorResponse) {
	return http.StatusInternalServerError, ErrorResponse{
		Code:    diagnostic.CodeInternalError,
		Message: "Внутренняя ошибка сервера",
		Details: ErrorDetails{
			Stage:     diagnostic.StageInternal,
			Retryable: false,
		},
	}
}
