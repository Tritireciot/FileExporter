package transformer

import "fmt"

type TemplateNotFoundError struct {
	id int
}

func (err *TemplateNotFoundError) Error() string {
	return fmt.Sprintf("Шаблон не неайден: %d", err.id)
}