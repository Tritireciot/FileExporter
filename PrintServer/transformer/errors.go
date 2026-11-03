package transformer

import "fmt"

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
