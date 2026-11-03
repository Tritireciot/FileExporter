package template_compiler

import (
	"PrintServer/PrintServer/diagnostic"
	"fmt"
	"strconv"
	"strings"
)

func scanConditionString(c *cursor) (raw string, diag *diagnostic.Diagnostic) {
	quote, ok := c.current()
	if !ok || (quote != '"' && quote != '\'') {
		panic("scanConditionString: курсор должен стоять на кавычке")
	}

	start := c.position()
	startIndex := c.runeIndex()
	c.advance(1)

	for !c.eof() {
		current, _ := c.current()

		if current == '\n' {
			break
		}

		if current == '\\' {
			c.advance(1)
			if c.eof() {
				break
			}
			next, _ := c.current()
			if next == '\n' {
				break
			}
			c.advance(1)
			continue
		}

		c.advance(1)
		if current == quote {
			return c.slice(startIndex, c.runeIndex()), nil
		}
	}

	return "", diagnostic.NewTranspile(
		diagnostic.CodeUnclosedString,
		"Строковый литерал в условии не закрыт",
	).
		WithHint("Закройте строку такой же кавычкой").
		WithSpan(start, c.position())
}

func normalizeStringLiteral(raw string) (string, error) {
	runes := []rune(raw)
	if len(runes) < 2 {
		return "", fmt.Errorf("строковый литерал не закрыт")
	}
	quote := runes[0]
	if (quote != '\'' && quote != '"') || runes[len(runes)-1] != quote {
		return "", fmt.Errorf("строковый литерал должен быть окружён одинаковыми кавычками")
	}

	var decoded strings.Builder
	content := runes[1 : len(runes)-1]
	for i := 0; i < len(content); i++ {
		current := content[i]
		if current == '\n' {
			return "", fmt.Errorf("перевод строки внутри литерала не допускается")
		}

		if current != '\\' {
			decoded.WriteRune(current)
			continue
		}

		if i+1 >= len(content) {
			return "", fmt.Errorf("обратный слеш в конце строкового литерала")
		}

		i++

		switch content[i] {
		case '\\':
			decoded.WriteRune('\\')
		case '\'':
			decoded.WriteRune('\'')
		case '"':
			decoded.WriteRune('"')
		case 'n':
			decoded.WriteRune('\n')
		case 't':
			decoded.WriteRune('\t')
		case 'r':
			decoded.WriteRune('\r')
		default:
			return "", fmt.Errorf("неизвестная escape-последовательность \\%c", content[i])
		}
	}
	return strconv.Quote(decoded.String()), nil
}
