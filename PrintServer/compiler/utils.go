package template_compiler

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

func isValidFieldChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) ||
		strings.ContainsRune("._| :#()\\$[]=<>!&\"'+-*/", r)
}

func isValidRepeatPathRune(r rune) bool {
	return unicode.IsLetter(r) ||
		unicode.IsDigit(r) ||
		strings.ContainsRune("._[]$", r)
}

func isConditionOperandStart(r rune) bool {
	return unicode.IsLetter(r) ||
		unicode.IsDigit(r) ||
		r == '_' ||
		r == '.' ||
		r == '$'
}

func isConditionOperandPart(r rune) bool {
	return isConditionOperandStart(r) ||
		r == '[' ||
		r == ']'
}

func isPathIdentifierStart(r rune) bool {
	return unicode.IsLetter(r) || r == '_'
}

func isPathIdentifierPart(r rune) bool {
	return isPathIdentifierStart(r) ||
		unicode.IsDigit(r)
}

func isValidFieldPath(path string) bool {
	if path == "" {
		return false
	}

	expectIdentifierStart := true

	for _, current := range path {
		if current == '.' {
			if expectIdentifierStart {
				return false
			}

			expectIdentifierStart = true
			continue
		}

		if expectIdentifierStart {
			if !isPathIdentifierStart(current) {
				return false
			}

			expectIdentifierStart = false
			continue
		}

		if !isPathIdentifierPart(current) {
			return false
		}
	}

	return !expectIdentifierStart
}

func isValidScopeAlias(alias string) bool {
	if alias == "" {
		return false
	}

	runes := []rune(alias)

	if !isPathIdentifierStart(runes[0]) {
		return false
	}

	for _, current := range runes[1:] {
		if !isPathIdentifierPart(current) {
			return false
		}
	}

	return true
}

func isQuotedLiteral(value string) bool {
	if len(value) < 2 {
		return false
	}

	first := value[0]
	last := value[len(value)-1]

	switch first {
	case '"':
		return last == '"'

	case '\'':
		return last == '\''

	case '`':
		return last == '`'

	default:
		return false
	}
}

func validateIntegerLiteral(value string) (string, error) {
	if value == "" {
		return "", fmt.Errorf("пустая строка не целое число")
	}

	start := 0
	if value[0] == '-' || value[0] == '+' {
		if len(value) == 1 {
			return "", fmt.Errorf("целое число не может состоять только из знака")
		}
		start = 1
	}

	for index := start; index < len(value); index++ {
		if value[index] < '0' || value[index] > '9' {
			return "", fmt.Errorf("целое число состоит только из символов от 0 до 9")
		}
	}

	n, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return "", fmt.Errorf("целое число вне диапазона int64 или записано неверно: %w", err)
	}
	if n < -(1<<53) || n > 1<<53 {
		return "", fmt.Errorf("целое число вне допустимого диапазона")
	}
	return strconv.FormatInt(n, 10), nil
}

func validateFloatLiteral(value string) (string, error) {
	if value == "" {
		return "", fmt.Errorf("пустая строка не число c плавающей точкой")
	}

	start := 0
	if value[0] == '-' || value[0] == '+' {
		if len(value) == 1 {
			return "", fmt.Errorf("число c плавающей точкой не может состоять только из знака")
		}
		start = 1
	}

	dotSeen := false
	digitSeen := false

	for index := start; index < len(value); index++ {
		switch {
		case value[index] >= '0' && value[index] <= '9':
			digitSeen = true

		case value[index] == '.' && !dotSeen:
			dotSeen = true

		default:
			return "", fmt.Errorf("число c плавающей точкой состоит только из символов от 0 до 9 и '.'")
		}
	}

	if !(digitSeen && dotSeen) {
		return "", fmt.Errorf("число c плавающей точкой не может быть без точки или без цифр")
	}
	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return "", fmt.Errorf("число c плавающей точкой вне диапазона float64 или записано неверно: %w", err)
	}

	normalized := strconv.FormatFloat(f, 'g', -1, 64)
	if !strings.ContainsAny(normalized, ".eE") {
		normalized += ".0"
	}
	return normalized, nil
}

func isFieldLiteral(value string) bool {
	switch value {
	case "true", "false", "nil":
		return true
	}

	if isQuotedLiteral(value) {
		return true
	}

	if _, err := validateIntegerLiteral(value); err == nil {
		return true
	}

	if _, err := validateFloatLiteral(value); err == nil {
		return true
	}

	return false
}

func isValidFormatterName(value string) bool {
	if value == "" {
		return false
	}

	for index, current := range value {
		if index == 0 {
			if current != '_' && !unicode.IsLetter(current) {
				return false
			}
			continue
		}

		if current != '_' &&
			!unicode.IsLetter(current) &&
			!unicode.IsDigit(current) {
			return false
		}
	}

	return true
}
