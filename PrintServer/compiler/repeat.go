package template_compiler

import (
	"PrintServer/PrintServer/diagnostic"
	"fmt"
)

func (s *parseSession) parseRepeat() *diagnostic.Diagnostic {
	const prefix = "<#Repeat#"
	const suffix = "#>"

	startPosition := s.cursor.position()

	s.cursor.advance(len(prefix))

	fieldStartPosition := s.cursor.position()
	fieldStartIndex := s.cursor.runeIndex()

	for !s.cursor.eof() && !s.cursor.startsWith(suffix) {
		s.cursor.advance(1)
	}

	if s.cursor.eof() {
		return diagnostic.NewTranspile(
			diagnostic.CodeUnclosedSpecialTag,
			"Конструкция Repeat не закрыта",
		).
			WithHint("Добавьте закрывающую последовательность #>").
			WithSpan(
				startPosition,
				s.cursor.position(),
			)
	}

	fieldEndPosition := s.cursor.position()
	fieldEndIndex := s.cursor.runeIndex()
	rawFieldName := s.cursor.slice(fieldStartIndex, fieldEndIndex)

	if rawFieldName == "" {
		return diagnostic.NewTranspile(
			diagnostic.CodeInvalidRepeat,
			"В конструкции Repeat не указан путь",
		).
			WithHint("Укажите путь к массиву, например <#Repeat#Monitoring.Items#>").
			WithSpan(
				startPosition,
				s.cursor.position(),
			)
	}

	rangePath, err := s.compilePath(
		rawFieldName,
		diagnostic.Span{
			Start: fieldStartPosition,
			End:   fieldEndPosition,
		},
		s.buildContextPath,
	)
	if err != nil {
		return err
	}

	s.cursor.advance(len(suffix))
	cleanName := cleanVarName(rawFieldName)

	s.encounter++
	variableID := s.encounter

	contextVar := fmt.Sprintf(
		"$%s%d",
		cleanName,
		variableID,
	)

	contextIndex := fmt.Sprintf(
		"$%sIndex%d",
		cleanName,
		variableID,
	)

	userIndexName := cleanName + "Index"

	s.blockStack = append(s.blockStack, Scope{
		BlockType:     blockTypeRepeat,
		OpenedAt:      startPosition,
		VarName:       contextVar,
		IndexName:     contextIndex,
		UserIndexName: userIndexName,
		Fields: map[string]string{
			cleanName: contextVar,
		},
	})

	s.output.WriteString(
		fmt.Sprintf(
			"{{ range %s, %s := %s }}",
			contextIndex,
			contextVar,
			rangePath,
		),
	)

	return nil
}
