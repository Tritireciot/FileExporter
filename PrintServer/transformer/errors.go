package transformer

import "fmt"

type TemplateIntegrityError struct {
	Expected string
	Actual   string
}

func (err *TemplateIntegrityError) Error() string {
	return fmt.Sprintf("published template content hash mismatch: expected %s, got %s", err.Expected, err.Actual)
}

type TemplateNotFoundError struct {
	id int
}

func (err *TemplateNotFoundError) Error() string {
	return fmt.Sprintf("Шаблон не неайден: %d", err.id)
}

type TemplateTypeNotFoundError struct {
	templateType string
}

func (err *TemplateTypeNotFoundError) Error() string {
	return fmt.Sprintf("Тип шаблона не неайден: %s", err.templateType)
}
