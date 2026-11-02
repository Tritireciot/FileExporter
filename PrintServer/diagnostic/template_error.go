package diagnostic

import (
	"fmt"
	"strings"
)

type TemplateError struct {
	Diagnostics []Diagnostic
	Cause       error
}

func (e *TemplateError) Error() string {
	if len(e.Diagnostics) == 0 {
		if e.Cause != nil {
			return e.Cause.Error()
		}

		return "ошибка обработки шаблона"
	}

	if len(e.Diagnostics) == 1 {
		if e.Cause != nil {
			return fmt.Sprintf(
				"%s: %v",
				e.Diagnostics[0].Message,
				e.Cause,
			)
		}

		return e.Diagnostics[0].Message
	}

	var builder strings.Builder

	fmt.Fprintf(
		&builder,
		"обнаружено ошибок шаблона: %d",
		len(e.Diagnostics),
	)

	if e.Cause != nil {
		fmt.Fprintf(&builder, ": %v", e.Cause)
	}

	return builder.String()
}

func (e *TemplateError) Unwrap() error {
	return e.Cause
}

func NewTemplateError(
	items []Diagnostic,
	cause error,
) *TemplateError {
	return &TemplateError{
		Diagnostics: items,
		Cause:       cause,
	}
}

func FromDiagnostic(
	item *Diagnostic,
	cause error,
) *TemplateError {
	if item == nil {
		return &TemplateError{
			Cause: cause,
		}
	}

	return &TemplateError{
		Diagnostics: []Diagnostic{*item},
		Cause:       cause,
	}
}
