package template_compiler

import (
	"PrintServer/PrintServer/diagnostic"
	"fmt"
	"strings"
	"unicode"
)

func (s *parseSession) resolveConditionBase(path string) string {
	resolved := s.resolveScopeReference(path)

	if strings.HasPrefix(resolved, "$") ||
		strings.HasPrefix(resolved, ".") {
		return resolved
	}

	return "." + resolved
}

func isDecimalIndex(value string) bool {
	if value == "" {
		return false
	}

	for _, current := range value {
		if current < '0' || current > '9' {
			return false
		}
	}

	return true
}

func (s *parseSession) compileIndexValue(
	value string,
	span diagnostic.Span,
) (string, *diagnostic.Diagnostic) {
	if value == "" {
		return "", &diagnostic.Diagnostic{
			Code:    diagnostic.CodeEmptyIndex,
			Stage:   diagnostic.StageTranspile,
			Message: "Индекс не содержит значения",
			Hint:    "Укажите номер элемента или имя переменной индекса",
			Span:    &span,
		}
	}

	if isDecimalIndex(value) {
		return value, nil
	}

	if !isValidFieldPath(value) {
		return "", &diagnostic.Diagnostic{
			Code:    diagnostic.CodeInvalidIndex,
			Stage:   diagnostic.StageTranspile,
			Message: "Индекс имеет неправильный формат",
			Hint:    "Используйте неотрицательное число или имя поля индекса",
			Span:    &span,
		}
	}

	return s.resolveConditionBase(value), nil
}

func (s *parseSession) compilePath(
	rawPath string,
	span diagnostic.Span,
	resolveBase func(string) string,
) (string, *diagnostic.Diagnostic) {
	if strings.TrimSpace(rawPath) == "" {
		return "", diagnostic.NewTranspile(
			diagnostic.CodeInvalidPath,
			"Путь к полю не указан",
		).
			WithHint("Укажите имя поля").
			WithSpan(
				span.Start,
				span.End,
			)
	}

	pathCursor := newCursorAt(rawPath, span.Start)

	baseStartIndex := pathCursor.runeIndex()

	for !pathCursor.eof() {
		current, _ := pathCursor.current()

		if current == '[' {
			break
		}

		if current == ']' {
			position := pathCursor.position()

			return "", diagnostic.NewTranspile(
				diagnostic.CodeInvalidPath,
				"Закрывающая квадратная скобка не имеет открывающей",
			).
				WithHint("Удалите лишнюю скобку или добавьте открывающую").
				WithSpan(
					position,
					position,
				)
		}

		pathCursor.advance(1)
	}

	baseEndIndex := pathCursor.runeIndex()
	basePath := pathCursor.slice(
		baseStartIndex,
		baseEndIndex,
	)

	if !isValidFieldPath(basePath) {
		return "", diagnostic.NewTranspile(
			diagnostic.CodeInvalidPath,
			"Начальная часть пути имеет неправильный формат",
		).
			WithHint("Используйте имена полей, разделённые точками").
			WithSpan(
				span.Start,
				pathCursor.position(),
			)
	}

	result := resolveBase(basePath)

	for !pathCursor.eof() {
		current, _ := pathCursor.current()

		switch current {
		case '[':
			openingPosition := pathCursor.position()
			pathCursor.advance(1)

			indexStartPosition := pathCursor.position()
			indexStartIndex := pathCursor.runeIndex()

			for !pathCursor.eof() {
				current, _ = pathCursor.current()

				if current == ']' {
					break
				}

				if current == '[' {
					position := pathCursor.position()

					return "", diagnostic.NewTranspile(
						diagnostic.CodeInvalidIndex,
						"Вложенные квадратные скобки в индексе не поддерживаются",
					).
						WithHint("Закройте текущий индекс перед открытием следующего").
						WithSpan(
							position,
							position,
						)
				}

				if unicode.IsSpace(current) {
					position := pathCursor.position()

					return "", diagnostic.NewTranspile(
						diagnostic.CodeInvalidIndex,
						"Пробелы внутри индекса не поддерживаются",
					).
						WithHint("Запишите индекс без пробелов, например Items[0]").
						WithSpan(
							position,
							position,
						)
				}

				pathCursor.advance(1)
			}

			if pathCursor.eof() {
				return "", diagnostic.NewTranspile(
					diagnostic.CodeUnclosedIndex,
					"Индекс не закрыт",
				).
					WithHint("Добавьте закрывающую квадратную скобку").
					WithSpan(
						openingPosition,
						pathCursor.position(),
					)
			}

			indexEndPosition := pathCursor.position()
			indexEndIndex := pathCursor.runeIndex()

			indexValue := pathCursor.slice(
				indexStartIndex,
				indexEndIndex,
			)

			compiledIndex, err := s.compileIndexValue(
				indexValue,
				diagnostic.Span{
					Start: indexStartPosition,
					End:   indexEndPosition,
				},
			)
			if err != nil {
				return "", err
			}

			// Пропускаем ']'.
			pathCursor.advance(1)

			result = fmt.Sprintf(
				"(index %s %s)",
				result,
				compiledIndex,
			)

		case '.':
			dotPosition := pathCursor.position()
			pathCursor.advance(1)

			if pathCursor.eof() {
				return "", diagnostic.NewTranspile(
					diagnostic.CodeInvalidPath,
					"После точки отсутствует имя поля",
				).
					WithHint("Добавьте имя поля после точки").
					WithSpan(
						dotPosition,
						pathCursor.position(),
					)
			}

			fieldStartPosition := pathCursor.position()
			fieldStartIndex := pathCursor.runeIndex()

			current, _ = pathCursor.current()

			if !isPathIdentifierStart(current) {
				return "", diagnostic.NewTranspile(
					diagnostic.CodeInvalidPath,
					"После точки указано неправильное имя поля",
				).
					WithHint("Имя поля должно начинаться с буквы или символа подчёркивания").
					WithSpan(
						fieldStartPosition,
						fieldStartPosition,
					)
			}

			pathCursor.advance(1)

			for !pathCursor.eof() {
				current, _ = pathCursor.current()

				if !isPathIdentifierPart(current) {
					break
				}

				pathCursor.advance(1)
			}

			fieldName := pathCursor.slice(
				fieldStartIndex,
				pathCursor.runeIndex(),
			)

			result += "." + fieldName

		default:
			position := pathCursor.position()
			pathCursor.advance(1)

			return "", diagnostic.NewTranspile(
				diagnostic.CodeInvalidPath,
				fmt.Sprintf(
					"Недопустимый символ %q после элемента пути",
					current,
				),
			).
				WithHint("После элемента пути ожидается точка, индекс или конец пути").
				WithSpan(
					position,
					pathCursor.position(),
				)
		}
	}

	return result, nil
}
