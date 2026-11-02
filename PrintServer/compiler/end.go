package template_compiler

import "PrintServer/PrintServer/diagnostic"

func (s *parseSession) parseEnd() *diagnostic.Diagnostic {
	const endTag = "<#End#>"

	startPosition := s.cursor.position()

	s.cursor.advance(len(endTag))

	if len(s.blockStack) == 0 {
		return diagnostic.NewTranspile(
			diagnostic.CodeUnexpectedEnd,
			"Обнаружен End без открытого блока",
		).
			WithHint("Удалите лишний <#End#> или добавьте открывающую конструкцию").
			WithSpan(
				startPosition,
				s.cursor.position(),
			)
	}

	lastIndex := len(s.blockStack) - 1
	s.blockStack = s.blockStack[:lastIndex]

	s.output.WriteString("{{ end }}")

	return nil
}
