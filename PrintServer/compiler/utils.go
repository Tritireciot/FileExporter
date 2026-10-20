package template_compiler

import (
	"strings"
	"unicode"
)

func isValidFieldChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) ||
		strings.ContainsRune("._| :#()\\$[]=<>!&\"'+-*/", r)
}
