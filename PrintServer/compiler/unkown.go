package template_compiler

import (
	"PrintServer/PrintServer/diagnostic"
	"fmt"
)

func (s *parseSession) parseUnknownSpecialTag() *diagnostic.Diagnostic {
	const prefix = "<#"
	const suffix = "#>"

	startPosition := s.cursor.position()

	s.cursor.advance(len(prefix))

	unkownTagStartIndex := s.cursor.runeIndex()

	for !s.cursor.eof() && !s.cursor.startsWith(suffix) {
		s.cursor.advance(1)
	}

	fieldEndIndex := s.cursor.runeIndex()
	unkownTagName := s.cursor.slice(unkownTagStartIndex, fieldEndIndex)

	return diagnostic.NewTranspile(
		diagnostic.CodeUnknownSpecialTag,
		fmt.Sprintf("Неизвестный специальный тег: %s", unkownTagName),
	).WithSpan(
		startPosition,
		s.cursor.position(),
	)
}
