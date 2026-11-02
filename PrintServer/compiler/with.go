package template_compiler

import (
	"PrintServer/PrintServer/diagnostic"
	"fmt"
)

func (s *parseSession) parseWith() *diagnostic.Diagnostic {
	const prefix = "<#With#"
	const suffix = "#>"

	startPosition := s.cursor.position()

	s.cursor.advance(len(prefix))

	bodyStartPosition := s.cursor.position()
	bodyStartIndex := s.cursor.runeIndex()

	for !s.cursor.eof() && !s.cursor.startsWith(suffix) {
		s.cursor.advance(1)
	}

	if s.cursor.eof() {
		return diagnostic.NewTranspile(
			diagnostic.CodeUnclosedSpecialTag,
			"Конструкция With не закрыта",
		).
			WithHint("Добавьте закрывающую последовательность #>").
			WithSpan(
				startPosition,
				s.cursor.position(),
			)
	}

	bodyEndPosition := s.cursor.position()
	bodyEndIndex := s.cursor.runeIndex()

	body := s.cursor.slice(
		bodyStartIndex,
		bodyEndIndex,
	)

	bodyRunes := []rune(body)

	separatorIndex := -1

	for index, current := range bodyRunes {
		if current != '#' {
			continue
		}

		if separatorIndex != -1 {
			return diagnostic.NewTranspile(
				diagnostic.CodeInvalidWith,
				"Конструкция With содержит лишний разделитель #",
			).
				WithHint("Используйте формат <#With#alias#path#>").
				WithSpan(
					bodyStartPosition,
					bodyEndPosition,
				)
		}

		separatorIndex = index
	}

	if separatorIndex == -1 {
		return diagnostic.NewTranspile(
			diagnostic.CodeInvalidWith,
			"В конструкции With отсутствует путь",
		).
			WithHint("Используйте формат <#With#alias#path#>").
			WithSpan(
				bodyStartPosition,
				bodyEndPosition,
			)
	}

	alias := string(bodyRunes[:separatorIndex])
	rawPath := string(bodyRunes[separatorIndex+1:])

	if !isValidScopeAlias(alias) {
		return diagnostic.NewTranspile(
			diagnostic.CodeInvalidWith,
			"Псевдоним With имеет неправильный формат",
		).
			WithHint("Используйте одно имя из букв, цифр и символов подчёркивания").
			WithSpan(
				bodyStartPosition,
				bodyEndPosition,
			)
	}

	if rawPath == "" {
		return diagnostic.NewTranspile(
			diagnostic.CodeInvalidWith,
			"В конструкции With не указан путь",
		).
			WithHint("Укажите путь, например <#With#equipment#Equipments[0]#>").
			WithSpan(
				bodyStartPosition,
				bodyEndPosition,
			)
	}

	// Вычисляем позицию начала пути.
	bodyCursor := newCursorAt(body, bodyStartPosition)
	bodyCursor.advance(separatorIndex + 1)

	pathStartPosition := bodyCursor.position()

	withPath, err := s.compilePath(
		rawPath,
		diagnostic.Span{
			Start: pathStartPosition,
			End:   bodyEndPosition,
		},
		s.buildContextPath,
	)
	if err != nil {
		return err
	}

	s.cursor.advance(len(suffix))

	s.encounter++
	contextVar := fmt.Sprintf(
		"$with%d",
		s.encounter,
	)

	s.blockStack = append(
		s.blockStack,
		Scope{
			BlockType: blockTypeWith,
			OpenedAt:  startPosition,
			VarName:   contextVar,
			Fields: map[string]string{
				alias: contextVar,
			},
		},
	)

	s.output.WriteString(
		fmt.Sprintf(
			"{{ with %s := %s }}",
			contextVar,
			withPath,
		),
	)

	return nil
}
