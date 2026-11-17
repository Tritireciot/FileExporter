package template_compiler

import (
	"PrintServer/PrintServer/diagnostic"
	"fmt"
	"strings"
	"unicode"
)

func isHrefFragment(c *cursor) bool {
	tagStart := -1
	for i := c.index - 1; i >= 0; i-- {
		if c.source[i] == '>' {
			return false
		}
		if c.source[i] == '<' {
			tagStart = i + 1
			break
		}
	}
	if tagStart < 0 {
		return false
	}

	for i := tagStart; i+4 <= c.index; i++ {
		if i > tagStart && !unicode.IsSpace(c.source[i-1]) {
			continue
		}
		if !strings.EqualFold(string(c.source[i:i+4]), "href") {
			continue
		}

		j := i + 4
		for j < c.index && unicode.IsSpace(c.source[j]) {
			j++
		}
		if j >= c.index || c.source[j] != '=' {
			continue
		}
		j++
		for j < c.index && unicode.IsSpace(c.source[j]) {
			j++
		}
		quote := rune(0)
		if j < c.index && (c.source[j] == '"' || c.source[j] == '\'') {
			quote = c.source[j]
			j++
		}
		for k := j; k < c.index; k++ {
			if c.source[k] == quote || (quote == 0 && (unicode.IsSpace(c.source[k]) || c.source[k] == '>')) {
				return false
			}
		}

		for k := c.index + 1; k < len(c.source); k++ {
			r := c.source[k]
			if r == '#' {
				return false
			}
			if r == quote || (quote == 0 && (unicode.IsSpace(r) || r == '>')) {
				return true
			}
		}
		return false
	}
	return false
}

func cssIDSelectorLength(c *cursor) int {
	first, ok := c.peek(1)
	if !ok || !(unicode.IsLetter(first) || first == '_' || first == '-') {
		return 0
	}

	length := 2
	for {
		r, ok := c.peek(length)
		if !ok || !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-') {
			break
		}
		length++
	}

	next, ok := c.peek(length)
	if !ok {
		return 0
	}
	if next == '#' {
		afterHash, ok := c.peek(length + 1)
		if !ok || !(unicode.IsLetter(afterHash) || afterHash == '_' || afterHash == '-') {
			return 0
		}
	} else if !unicode.IsSpace(next) && !strings.ContainsRune("{.,:[]>+~", next) {
		return 0
	}

	for offset := length; ; offset++ {
		r, ok := c.peek(offset)
		if !ok || r == ';' || r == '}' || r == '<' {
			return 0
		}
		if r == '{' {
			return length
		}
	}
}

func cssHexColorLength(c *cursor) int {
	length := 0
	for {
		r, ok := c.peek(length + 1)
		if !ok || !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			break
		}
		length++
	}

	if length != 3 && length != 4 && length != 6 && length != 8 {
		return 0
	}

	next, ok := c.peek(length + 1)
	if ok && !unicode.IsSpace(next) && !strings.ContainsRune(";},)'\"<!", next) {
		return 0
	}
	return length + 1 // leading #
}

func cleanVarName(path string) string {
	if idx := strings.LastIndex(path, "."); idx != -1 {
		path = path[idx+1:]
	}
	if idx := strings.Index(path, "["); idx != -1 {
		path = path[:idx]
	}
	return path
}

func (s *parseSession) tryParseField() (bool, *diagnostic.Diagnostic) {
	start := s.cursor.position()

	probe := s.cursor
	probe.advance(1) // открывающий #
	unclosedField := func() (bool, *diagnostic.Diagnostic) {
		return true, diagnostic.NewTranspile(
			diagnostic.CodeUnclosedField,
			"Поле шаблона не закрыто символом #",
		).WithHint("Добавьте закрывающий символ #").WithSpan(start, probe.position())
	}

	contentStart := probe.runeIndex()

	var quote rune
	escaped := false

	for !probe.eof() {
		current, _ := probe.current()

		if quote != 0 {
			if escaped {
				escaped = false
				probe.advance(1)
				continue
			}

			if current == '\\' {
				escaped = true
				probe.advance(1)
				continue
			}

			if current == quote {
				quote = 0
			}

			probe.advance(1)
			continue
		}

		switch current {
		case '"', '\'', '`':
			quote = current
			probe.advance(1)

		case '#':
			rawField := probe.slice(contentStart, probe.runeIndex())

			s.cursor = probe
			s.cursor.advance(1)

			span := diagnostic.Span{
				Start: start,
				End:   s.cursor.position(),
			}

			expression, err := s.compileFieldExpression(rawField, span)
			if err != nil {
				return true, err
			}

			s.output.WriteString("{{ ")
			s.output.WriteString(expression)
			s.output.WriteString(" }}")

			return true, nil

		case '\n', '\r', '<', '>', '{', '}', ';':
			return unclosedField()

		default:
			probe.advance(1)
		}
	}

	return unclosedField()
}

func (s *parseSession) compileFieldExpression(
	raw string,
	span diagnostic.Span,
) (string, *diagnostic.Diagnostic) {
	parts, err := splitFieldPipeline(raw, span)
	if err != nil {
		return "", err
	}

	if len(parts) == 0 || strings.TrimSpace(parts[0]) == "" {
		return "", diagnostic.NewTranspile(
			diagnostic.CodeEmptyField,
			"Поле между символами # не указано",
		).WithSpan(
			span.Start,
			span.End,
		)
	}

	baseTokens, err := tokenizeFieldCommand(parts[0], span)
	if err != nil {
		return "", err
	}

	baseExpression, compileErr := s.compileInitialFieldCommand(
		baseTokens,
		span,
	)
	if compileErr != nil {
		return "", compileErr
	}

	var result strings.Builder
	result.WriteString(baseExpression)

	for _, rawCommand := range parts[1:] {
		commandTokens, commandErr := tokenizeFieldCommand(rawCommand, span)
		if commandErr != nil {
			return "", commandErr
		}

		if len(commandTokens) == 0 {
			return "", diagnostic.NewTranspile(
				diagnostic.CodeInvalidFieldPipeline,
				"После символа | должен быть указан форматтер",
			).WithSpan(
				span.Start,
				span.End,
			)
		}

		formatter := commandTokens[0]
		if !isValidFormatterName(formatter) {
			return "", diagnostic.NewTranspile(
				diagnostic.CodeInvalidFormatter,
				fmt.Sprintf("Некорректное имя форматтера %q", formatter),
			).WithSpan(
				span.Start,
				span.End,
			)
		}

		result.WriteString(" | ")
		result.WriteString(formatter)

		for _, rawArgument := range commandTokens[1:] {
			argument, argumentErr := s.compileFieldArgument(
				rawArgument,
				span,
			)
			if argumentErr != nil {
				return "", argumentErr
			}

			result.WriteByte(' ')
			result.WriteString(argument)
		}
	}

	return result.String(), nil
}

func splitFieldPipeline(
	raw string,
	span diagnostic.Span,
) ([]string, *diagnostic.Diagnostic) {
	var result []string
	var current strings.Builder

	var quote rune
	escaped := false

	for _, char := range raw {
		if quote != 0 {
			current.WriteRune(char)

			if escaped {
				escaped = false
				continue
			}

			if char == '\\' && quote != '`' {
				escaped = true
				continue
			}

			if char == quote {
				quote = 0
			}

			continue
		}

		switch char {
		case '"', '\'', '`':
			quote = char
			current.WriteRune(char)

		case '|':
			result = append(result, strings.TrimSpace(current.String()))
			current.Reset()

		default:
			current.WriteRune(char)
		}
	}

	if quote != 0 {
		return nil, diagnostic.NewTranspile(
			diagnostic.CodeInvalidFieldPipeline,
			"В выражении поля не закрыт строковый литерал",
		).WithSpan(
			span.Start,
			span.End,
		)
	}

	result = append(result, strings.TrimSpace(current.String()))
	return result, nil
}

func tokenizeFieldCommand(
	raw string,
	span diagnostic.Span,
) ([]string, *diagnostic.Diagnostic) {
	var tokens []string
	var current strings.Builder

	var quote rune
	escaped := false

	flush := func() {
		if current.Len() == 0 {
			return
		}

		tokens = append(tokens, current.String())
		current.Reset()
	}

	for _, char := range raw {
		if quote != 0 {
			current.WriteRune(char)

			if escaped {
				escaped = false
				continue
			}

			if char == '\\' && quote != '`' {
				escaped = true
				continue
			}

			if char == quote {
				quote = 0
			}

			continue
		}

		switch {
		case char == '"' || char == '\'' || char == '`':
			quote = char
			current.WriteRune(char)

		case unicode.IsSpace(char):
			flush()

		default:
			current.WriteRune(char)
		}
	}

	if quote != 0 {
		return nil, diagnostic.NewTranspile(
			diagnostic.CodeInvalidFieldPipeline,
			"В аргументах форматтера не закрыт строковый литерал",
		).WithSpan(
			span.Start,
			span.End,
		)
	}

	flush()
	return tokens, nil
}

func (s *parseSession) compileFieldArgument(
	rawArgument string,
	span diagnostic.Span,
) (string, *diagnostic.Diagnostic) {
	argument := strings.TrimSpace(rawArgument)

	if argument == "" {
		return "", diagnostic.NewTranspile(
			diagnostic.CodeInvalidFieldPipeline,
			"Аргумент форматтера не может быть пустым",
		).WithSpan(
			span.Start,
			span.End,
		)
	}

	if isFieldLiteral(argument) {
		return argument, nil
	}

	if argument == "." {
		return argument, nil
	}

	return s.compilePath(
		argument,
		span,
		s.buildContextPath,
	)
}

func (s *parseSession) compileInitialFieldCommand(
	tokens []string,
	span diagnostic.Span,
) (string, *diagnostic.Diagnostic) {
	if len(tokens) == 0 {
		return "", diagnostic.NewTranspile(
			diagnostic.CodeEmptyField,
			"Выражение поля не может быть пустым",
		).WithSpan(
			span.Start,
			span.End,
		)
	}

	if len(tokens) == 1 {
		return s.compilePath(
			tokens[0],
			span,
			s.buildContextPath,
		)
	}

	functionName := tokens[0]

	if !isValidFormatterName(functionName) {
		return "", diagnostic.NewTranspile(
			diagnostic.CodeInvalidFormatter,
			fmt.Sprintf(
				"Некорректное имя функции %q",
				functionName,
			),
		).WithSpan(
			span.Start,
			span.End,
		)
	}

	var result strings.Builder
	result.WriteString(functionName)

	for _, rawArgument := range tokens[1:] {
		argument, err := s.compileFieldArgument(rawArgument, span)
		if err != nil {
			return "", err
		}

		result.WriteByte(' ')
		result.WriteString(argument)
	}

	return result.String(), nil
}
