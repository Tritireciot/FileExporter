package template_compiler

import (
	"PrintServer/PrintServer/diagnostic"
	"fmt"
	"strings"
)

func (s *parseSession) readCondition(
	prefix string,
	startPosition diagnostic.Location,
) (string, diagnostic.Location, *diagnostic.Diagnostic) {
	const suffix = "#>"

	s.cursor.advance(len(prefix))

	conditionPosition := s.cursor.position()
	conditionStart := s.cursor.runeIndex()

	for !s.cursor.eof() && !s.cursor.startsWith(suffix) {
		s.cursor.advance(1)
	}

	if s.cursor.eof() {
		return "", conditionPosition, &diagnostic.Diagnostic{
			Code:    diagnostic.CodeUnclosedSpecialTag,
			Stage:   diagnostic.StageTranspile,
			Message: "Условная конструкция не закрыта",
			Hint:    "Добавьте закрывающую последовательность #>",
			Span: &diagnostic.Span{
				Start: startPosition,
				End:   s.cursor.position(),
			},
		}
	}

	conditionEnd := s.cursor.runeIndex()
	condition := s.cursor.slice(conditionStart, conditionEnd)

	// Пропускаем "#>".
	s.cursor.advance(len(suffix))

	if strings.TrimSpace(condition) == "" {
		return "", conditionPosition, &diagnostic.Diagnostic{
			Code:    diagnostic.CodeExpectedOperand,
			Stage:   diagnostic.StageTranspile,
			Message: "В условной конструкции отсутствует условие",
			Hint:    "Укажите выражение, например <#If Equipment.Active#>",
			Span: &diagnostic.Span{
				Start: conditionPosition,
				End:   s.cursor.position(),
			},
		}
	}

	return condition, conditionPosition, nil
}

func (s *parseSession) parseIf() *diagnostic.Diagnostic {
	const prefix = "<#If "

	startPosition := s.cursor.position()

	condition, conditionPosition, err := s.readCondition(
		prefix,
		startPosition,
	)
	if err != nil {
		return err
	}

	parsedCondition, err := s.parseCondition(
		condition,
		conditionPosition,
	)
	if err != nil {
		return err
	}

	s.blockStack = append(s.blockStack, Scope{
		BlockType: blockTypeIf,
		OpenedAt:  startPosition,
	})

	s.output.WriteString(
		fmt.Sprintf("{{ if %s }}", parsedCondition),
	)

	return nil
}

func (s *parseSession) parseElseIf() *diagnostic.Diagnostic {
	const prefix = "<#ElseIf "

	startPosition := s.cursor.position()

	condition, conditionPosition, err := s.readCondition(
		prefix,
		startPosition,
	)
	if err != nil {
		return err
	}

	if len(s.blockStack) == 0 {
		return &diagnostic.Diagnostic{
			Code:    diagnostic.CodeElseIfWithoutIf,
			Stage:   diagnostic.StageTranspile,
			Message: "ElseIf используется без открытого If",
			Hint:    "Добавьте перед ElseIf конструкцию If",
			Span: &diagnostic.Span{
				Start: startPosition,
				End:   s.cursor.position(),
			},
		}
	}

	scopeIndex := len(s.blockStack) - 1
	scope := &s.blockStack[scopeIndex]

	if scope.BlockType != blockTypeIf {
		return &diagnostic.Diagnostic{
			Code:    diagnostic.CodeElseIfWithoutIf,
			Stage:   diagnostic.StageTranspile,
			Message: "ElseIf не относится к открытому If",
			Hint:    "Сначала закройте текущий вложенный блок через <#End#>",
			Span: &diagnostic.Span{
				Start: startPosition,
				End:   s.cursor.position(),
			},
		}
	}

	if scope.ElseSeen {
		return &diagnostic.Diagnostic{
			Code:    diagnostic.CodeElseIfAfterElse,
			Stage:   diagnostic.StageTranspile,
			Message: "ElseIf не может находиться после Else",
			Hint:    "Переместите ElseIf перед Else",
			Span: &diagnostic.Span{
				Start: startPosition,
				End:   s.cursor.position(),
			},
		}
	}

	parsedCondition, err := s.parseCondition(
		condition,
		conditionPosition,
	)
	if err != nil {
		return err
	}

	s.output.WriteString(
		fmt.Sprintf("{{ else if %s }}", parsedCondition),
	)

	return nil
}

func (s *parseSession) parseElse() *diagnostic.Diagnostic {
	const elseTag = "<#Else#>"

	startPosition := s.cursor.position()
	s.cursor.advance(len(elseTag))

	if len(s.blockStack) == 0 {
		return &diagnostic.Diagnostic{
			Code:    diagnostic.CodeElseWithoutIf,
			Stage:   diagnostic.StageTranspile,
			Message: "Else используется без открытого If",
			Hint:    "Добавьте перед Else конструкцию If",
			Span: &diagnostic.Span{
				Start: startPosition,
				End:   s.cursor.position(),
			},
		}
	}

	scopeIndex := len(s.blockStack) - 1
	scope := &s.blockStack[scopeIndex]

	if scope.BlockType != blockTypeIf {
		return &diagnostic.Diagnostic{
			Code:    diagnostic.CodeElseWithoutIf,
			Stage:   diagnostic.StageTranspile,
			Message: "Else не относится к открытому If",
			Hint:    "Сначала закройте текущий вложенный блок через <#End#>",
			Span: &diagnostic.Span{
				Start: startPosition,
				End:   s.cursor.position(),
			},
		}
	}

	if scope.ElseSeen {
		return &diagnostic.Diagnostic{
			Code:    diagnostic.CodeDuplicateElse,
			Stage:   diagnostic.StageTranspile,
			Message: "Конструкция If содержит несколько веток Else",
			Hint:    "Оставьте только одну ветку Else",
			Span: &diagnostic.Span{
				Start: startPosition,
				End:   s.cursor.position(),
			},
		}
	}

	scope.ElseSeen = true
	s.output.WriteString("{{ else }}")

	return nil
}
