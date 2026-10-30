package template_compiler

import (
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

func isIntegerLiteral(value string) bool {
	if value == "" {
		return false
	}

	start := 0
	if value[0] == '-' || value[0] == '+' {
		if len(value) == 1 {
			return false
		}
		start = 1
	}

	for index := start; index < len(value); index++ {
		if value[index] < '0' || value[index] > '9' {
			return false
		}
	}

	return true
}

func isFloatLiteral(value string) bool {
	if value == "" {
		return false
	}

	start := 0
	if value[0] == '-' || value[0] == '+' {
		if len(value) == 1 {
			return false
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
			return false
		}
	}

	return digitSeen && dotSeen
}

func isFieldLiteral(value string) bool {
	switch value {
	case "true", "false", "nil":
		return true
	}

	if isQuotedLiteral(value) {
		return true
	}

	if isIntegerLiteral(value) {
		return true
	}

	if isFloatLiteral(value) {
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
