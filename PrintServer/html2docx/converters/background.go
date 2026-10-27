package converters

import (
	"PrintServer/PrintServer/html2docx/css"
	"strings"

	"github.com/mmonterroca/docxgo/v2/domain"
)

func ConvertBackground(value css.StyleValue) *domain.Color {
	str := strings.TrimSpace(string(value))
	if str == "" {
		return nil
	}
	parts := strings.Fields(str)
	if len(parts) == 0 {
		return nil
	}
	return ConvertColor(parts[0])
}
