package template_compiler

import (
	"PrintServer/PrintServer/diagnostic"
	"fmt"
	"strings"
	"unicode"
)

type conditionTokenKind uint8

const (
	conditionTokenOperand conditionTokenKind = iota
	conditionTokenOperator
	conditionTokenLeftParenthesis
	conditionTokenRightParenthesis
)

type conditionToken struct {
	Kind  conditionTokenKind
	Value string
	Span  diagnostic.Span
}

var conditionTwoCharacterOperators = []string{
	"||",
	"&&",
	"==",
	"!=",
	"<=",
	">=",
}

const conditionSingleCharacterOperators = "+-*/<>!"

func tokenizeCondition(
	condition string,
	startPosition diagnostic.Location,
) ([]conditionToken, *diagnostic.Diagnostic) {
	conditionCursor := newCursorAt(condition, startPosition)
	tokens := make([]conditionToken, 0)

	for !conditionCursor.eof() {
		current, _ := conditionCursor.current()

		// Пробельные символы не являются токенами.
		if unicode.IsSpace(current) {
			conditionCursor.advance(1)
			continue
		}

		tokenStartPosition := conditionCursor.position()
		tokenStartIndex := conditionCursor.runeIndex()

		// Строковый литерал.
		if current == '"' || current == '\'' {
			quote := current
			conditionCursor.advance(1)

			closed := false

			for !conditionCursor.eof() {
				current, _ = conditionCursor.current()

				if current == '\n' {
					return nil, diagnostic.NewTranspile(
						diagnostic.CodeUnclosedCondition,
						"Строковый литерал в условии не закрыт",
					).
						WithHint("Закройте строковый литерал до конца строки").
						WithSpan(
							tokenStartPosition,
							conditionCursor.position(),
						)
				}

				// Экранированный символ внутри строки.
				if current == '\\' {
					conditionCursor.advance(1)

					if conditionCursor.eof() {
						break
					}

					conditionCursor.advance(1)
					continue
				}

				if current == quote {
					conditionCursor.advance(1)
					closed = true
					break
				}

				conditionCursor.advance(1)
			}

			if !closed {
				return nil, diagnostic.NewTranspile(
					diagnostic.CodeUnclosedCondition,
					"Строковый литерал в условии не закрыт",
				).
					WithHint("Добавьте закрывающую кавычку").
					WithSpan(
						tokenStartPosition,
						conditionCursor.position(),
					)
			}

			tokens = append(tokens, conditionToken{
				Kind: conditionTokenOperand,
				Value: conditionCursor.slice(
					tokenStartIndex,
					conditionCursor.runeIndex(),
				),
				Span: diagnostic.Span{
					Start: tokenStartPosition,
					End:   conditionCursor.position(),
				},
			})

			continue
		}

		// Открывающая скобка.
		if current == '(' {
			conditionCursor.advance(1)

			tokens = append(tokens, conditionToken{
				Kind:  conditionTokenLeftParenthesis,
				Value: "(",
				Span: diagnostic.Span{
					Start: tokenStartPosition,
					End:   conditionCursor.position(),
				},
			})

			continue
		}

		// Закрывающая скобка.
		if current == ')' {
			conditionCursor.advance(1)

			tokens = append(tokens, conditionToken{
				Kind:  conditionTokenRightParenthesis,
				Value: ")",
				Span: diagnostic.Span{
					Start: tokenStartPosition,
					End:   conditionCursor.position(),
				},
			})

			continue
		}

		// Сначала ищем двухсимвольный оператор.
		operator := ""

		for _, candidate := range conditionTwoCharacterOperators {
			if conditionCursor.startsWith(candidate) {
				operator = candidate
				break
			}
		}

		if operator != "" {
			conditionCursor.advance(len(operator))

			tokens = append(tokens, conditionToken{
				Kind:  conditionTokenOperator,
				Value: operator,
				Span: diagnostic.Span{
					Start: tokenStartPosition,
					End:   conditionCursor.position(),
				},
			})

			continue
		}

		// Затем односимвольный оператор.
		if strings.ContainsRune(
			conditionSingleCharacterOperators,
			current,
		) {
			conditionCursor.advance(1)

			tokens = append(tokens, conditionToken{
				Kind:  conditionTokenOperator,
				Value: string(current),
				Span: diagnostic.Span{
					Start: tokenStartPosition,
					End:   conditionCursor.position(),
				},
			})

			continue
		}

		// Поле, число, boolean или другой операнд.
		if isConditionOperandStart(current) {
			conditionCursor.advance(1)

			for !conditionCursor.eof() {
				current, _ = conditionCursor.current()

				if !isConditionOperandPart(current) {
					break
				}

				conditionCursor.advance(1)
			}

			tokens = append(tokens, conditionToken{
				Kind: conditionTokenOperand,
				Value: conditionCursor.slice(
					tokenStartIndex,
					conditionCursor.runeIndex(),
				),
				Span: diagnostic.Span{
					Start: tokenStartPosition,
					End:   conditionCursor.position(),
				},
			})

			continue
		}

		// Перед формированием ошибки двигаемся на один символ,
		// чтобы Span включал ошибочный символ.
		conditionCursor.advance(1)

		return nil, diagnostic.NewTranspile(
			diagnostic.CodeInvalidCondition,
			fmt.Sprintf(
				"Недопустимый символ %q в условии",
				current,
			),
		).
			WithHint("Используйте поля, литералы, скобки и поддерживаемые операторы").
			WithSpan(
				tokenStartPosition,
				conditionCursor.position(),
			)
	}

	return tokens, nil
}
