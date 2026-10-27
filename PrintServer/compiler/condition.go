package template_compiler

import (
	"PrintServer/PrintServer/diagnostic"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

var templateOperators = map[string]string{
	"||": "or",
	"&&": "and",
	"==": "eq",
	"!=": "ne",
	"<":  "lt",
	">":  "gt",
	"<=": "le",
	">=": "ge",
	"+":  "add",
	"-":  "sub",
	"*":  "mul",
	"/":  "div",
	"!":  "not",
}

func conditionOperatorPrecedence(operator string) int {
	switch operator {
	case "||":
		return 1

	case "&&":
		return 2

	case "==", "!=", "<", ">", "<=", ">=":
		return 3

	case "+", "-":
		return 4

	case "*", "/":
		return 5

	case "!":
		return 6

	default:
		return 0
	}
}

var conditionTemplateOperators = map[string]string{
	"||": "or",
	"&&": "and",
	"==": "eq",
	"!=": "ne",
	"<":  "lt",
	">":  "gt",
	"<=": "le",
	">=": "ge",
	"+":  "add",
	"-":  "sub",
	"*":  "mul",
	"/":  "div",
	"!":  "not",
}

func buildConditionPostfix(
	tokens []conditionToken,
) ([]conditionToken, *diagnostic.Diagnostic) {
	if len(tokens) == 0 {
		return nil, diagnostic.NewTranspile(
			diagnostic.CodeExpectedOperand,
			"Условие не содержит операндов",
		).
			WithHint("Укажите поле, литерал или логическое выражение")
	}

	output := make([]conditionToken, 0, len(tokens))
	operators := make([]conditionToken, 0)

	expectOperand := true

	for _, token := range tokens {
		switch token.Kind {
		case conditionTokenOperand:
			if !expectOperand {
				return nil, diagnostic.NewTranspile(
					diagnostic.CodeExpectedOperand,
					"Между операндами отсутствует оператор",
				).
					WithHint("Добавьте логический, сравнительный или арифметический оператор").
					WithSpan(
						token.Span.Start,
						token.Span.End,
					)
			}

			output = append(output, token)
			expectOperand = false

		case conditionTokenLeftParenthesis:
			if !expectOperand {
				return nil, diagnostic.NewTranspile(
					diagnostic.CodeExpectedOperand,
					"Перед открывающей скобкой отсутствует оператор",
				).
					WithHint("Добавьте оператор перед открывающей скобкой").
					WithSpan(
						token.Span.Start,
						token.Span.End,
					)
			}

			operators = append(operators, token)
			expectOperand = true

		case conditionTokenRightParenthesis:
			leftParenthesisIndex := -1

			for i := len(operators) - 1; i >= 0; i-- {
				if operators[i].Kind == conditionTokenLeftParenthesis {
					leftParenthesisIndex = i
					break
				}
			}

			if leftParenthesisIndex == -1 {
				return nil, diagnostic.NewTranspile(
					diagnostic.CodeUnmatchedParenthesis,
					"Закрывающая скобка не имеет открывающей",
				).
					WithHint("Добавьте открывающую скобку или удалите лишнюю закрывающую").
					WithSpan(
						token.Span.Start,
						token.Span.End,
					)
			}

			if expectOperand {
				return nil, diagnostic.NewTranspile(
					diagnostic.CodeExpectedOperand,
					"Перед закрывающей скобкой отсутствует операнд",
				).
					WithHint("Добавьте выражение внутри скобок").
					WithSpan(
						token.Span.Start,
						token.Span.End,
					)
			}

			for len(operators)-1 > leftParenthesisIndex {
				lastIndex := len(operators) - 1

				output = append(output, operators[lastIndex])
				operators = operators[:lastIndex]
			}

			operators = operators[:leftParenthesisIndex]

			expectOperand = false

		case conditionTokenOperator:
			precedence := conditionOperatorPrecedence(token.Value)

			if precedence == 0 {
				return nil, diagnostic.NewTranspile(
					diagnostic.CodeExpectedOperand,
					fmt.Sprintf(
						"Оператор %q не поддерживается",
						token.Value,
					),
				).
					WithHint("Используйте поддерживаемый оператор").
					WithSpan(
						token.Span.Start,
						token.Span.End,
					)
			}

			if token.Value == "!" {
				if !expectOperand {
					return nil, diagnostic.NewTranspile(
						diagnostic.CodeExpectedOperand,
						"Унарный оператор ! находится после операнда",
					).
						WithHint("Поместите ! перед операндом или выражением").
						WithSpan(
							token.Span.Start,
							token.Span.End,
						)
				}
			} else {
				if expectOperand {
					return nil, diagnostic.NewTranspile(
						diagnostic.CodeExpectedOperand,
						fmt.Sprintf(
							"Перед оператором %q отсутствует операнд",
							token.Value,
						),
					).
						WithHint("Добавьте операнд перед бинарным оператором").
						WithSpan(
							token.Span.Start,
							token.Span.End,
						)
				}
			}

			for len(operators) > 0 {
				lastIndex := len(operators) - 1
				top := operators[lastIndex]

				if top.Kind == conditionTokenLeftParenthesis {
					break
				}

				topPrecedence := conditionOperatorPrecedence(top.Value)

				shouldPop := false

				if token.Value == "!" {
					shouldPop = topPrecedence > precedence
				} else {
					shouldPop = topPrecedence >= precedence
				}

				if !shouldPop {
					break
				}

				output = append(output, top)
				operators = operators[:lastIndex]
			}

			operators = append(operators, token)
			expectOperand = true
		}
	}

	if expectOperand {
		lastToken := tokens[len(tokens)-1]

		return nil, diagnostic.NewTranspile(
			diagnostic.CodeExpectedOperand,
			"В конце условия отсутствует операнд",
		).
			WithHint("Добавьте операнд после оператора").
			WithSpan(
				lastToken.Span.Start,
				lastToken.Span.End,
			)
	}

	for len(operators) > 0 {
		lastIndex := len(operators) - 1
		operator := operators[lastIndex]
		operators = operators[:lastIndex]

		if operator.Kind == conditionTokenLeftParenthesis {
			return nil, diagnostic.NewTranspile(
				diagnostic.CodeUnmatchedParenthesis,
				"Открывающая скобка не имеет закрывающей",
			).
				WithHint("Добавьте закрывающую скобку").
				WithSpan(
					operator.Span.Start,
					operator.Span.End,
				)
		}

		output = append(output, operator)
	}

	return output, nil
}

func (s *parseSession) resolveConditionOperand(
	token conditionToken,
) (string, *diagnostic.Diagnostic) {
	value := token.Value

	if value == "" {
		return "", diagnostic.NewTranspile(
			diagnostic.CodeExpectedOperand,
			"Обнаружен пустой операнд",
		).
			WithSpan(
				token.Span.Start,
				token.Span.End,
			)
	}

	if strings.HasPrefix(value, `"`) ||
		strings.HasPrefix(value, `'`) {
		return value, nil
	}

	switch value {
	case "true", "false", "nil":
		return value, nil
	}

	firstRune, _ := utf8.DecodeRuneInString(value)

	if unicode.IsDigit(firstRune) {
		return value, nil
	}

	return s.compilePath(
		value,
		token.Span,
		s.resolveConditionBase,
	)
}

func (s *parseSession) buildConditionExpression(
	postfix []conditionToken,
) (string, *diagnostic.Diagnostic) {
	if len(postfix) == 0 {
		return "", diagnostic.NewTranspile(
			diagnostic.CodeExpectedOperand,
			"Условие не содержит выражения",
		).
			WithHint("Укажите поле, литерал или логическое выражение")
	}

	expressions := make([]string, 0, len(postfix))

	for _, token := range postfix {
		switch token.Kind {
		case conditionTokenOperand:
			operand, err := s.resolveConditionOperand(token)
			if err != nil {
				return "", err
			}

			expressions = append(expressions, operand)

		case conditionTokenOperator:
			templateOperator, ok :=
				conditionTemplateOperators[token.Value]

			if !ok {
				return "", diagnostic.NewTranspile(
					diagnostic.CodeExpectedOperand,
					fmt.Sprintf(
						"Оператор %q не поддерживается",
						token.Value,
					),
				).
					WithHint("Используйте поддерживаемый оператор").
					WithSpan(
						token.Span.Start,
						token.Span.End,
					)
			}

			if token.Value == "!" {
				if len(expressions) < 1 {
					return "", diagnostic.NewTranspile(
						diagnostic.CodeExpectedOperand,
						"Для оператора ! отсутствует операнд",
					).
						WithHint("Укажите выражение после оператора !").
						WithSpan(
							token.Span.Start,
							token.Span.End,
						)
				}

				lastIndex := len(expressions) - 1
				operand := expressions[lastIndex]

				expressions = expressions[:lastIndex]

				expressions = append(
					expressions,
					fmt.Sprintf(
						"(%s %s)",
						templateOperator,
						operand,
					),
				)

				continue
			}

			if len(expressions) < 2 {
				return "", diagnostic.NewTranspile(
					diagnostic.CodeExpectedOperand,
					fmt.Sprintf(
						"Для оператора %q недостаточно операндов",
						token.Value,
					),
				).
					WithHint("Бинарный оператор должен иметь левый и правый операнд").
					WithSpan(
						token.Span.Start,
						token.Span.End,
					)
			}

			// В postfix правый операнд находится на вершине стека.
			rightIndex := len(expressions) - 1
			right := expressions[rightIndex]

			leftIndex := rightIndex - 1
			left := expressions[leftIndex]

			// Убираем оба операнда.
			expressions = expressions[:leftIndex]

			expressions = append(
				expressions,
				fmt.Sprintf(
					"(%s %s %s)",
					templateOperator,
					left,
					right,
				),
			)
		}
	}

	if len(expressions) != 1 {
		lastToken := postfix[len(postfix)-1]

		return "", diagnostic.NewTranspile(
			diagnostic.CodeExpectedOperand,
			"Условие содержит несвязанные выражения",
		).
			WithHint("Добавьте оператор между выражениями").
			WithSpan(
				lastToken.Span.Start,
				lastToken.Span.End,
			)
	}

	return expressions[0], nil
}

func removeOuterParentheses(expression string) string {
	if len(expression) >= 2 &&
		expression[0] == '(' &&
		expression[len(expression)-1] == ')' {
		return expression[1 : len(expression)-1]
	}

	return expression
}

func (s *parseSession) parseCondition(
	condition string,
	startPosition diagnostic.Location,
) (string, *diagnostic.Diagnostic) {
	const legacyPrefix = "Condition:"

	// Оставляем поддержку старого синтаксиса.
	if strings.HasPrefix(condition, legacyPrefix) {
		condition = strings.TrimSpace(
			condition[len(legacyPrefix):],
		)
	}

	if condition == "" {
		return "", diagnostic.NewTranspile(
			diagnostic.CodeExpectedOperand,
			"Условие не содержит выражения",
		).
			WithHint("Укажите поле, литерал или логическое выражение").
			WithSpan(
				startPosition,
				startPosition,
			)
	}

	if strings.TrimSpace(condition) == "" {
		return "", diagnostic.NewTranspile(
			diagnostic.CodeExpectedOperand,
			"Условие не содержит выражения",
		).
			WithHint("Укажите поле, литерал или логическое выражение").
			WithSpan(
				startPosition,
				startPosition,
			)
	}

	tokens, err := tokenizeCondition(condition, startPosition)
	if err != nil {
		return "", err
	}

	postfix, err := buildConditionPostfix(tokens)
	if err != nil {
		return "", err
	}

	result, err := s.buildConditionExpression(postfix)
	if err != nil {
		return "", err
	}

	return removeOuterParentheses(result), nil
}
