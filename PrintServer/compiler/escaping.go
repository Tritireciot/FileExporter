package template_compiler

import (
	"PrintServer/PrintServer/diagnostic"
	"strconv"
)

func (s *parseSession) parseEscape() *diagnostic.Diagnostic {
	s.cursor.advance(1)

	if s.cursor.eof() {
		s.output.WriteRune('\\')
		return nil
	}

	current, _ := s.cursor.current()

	switch current {
	case '#':
		s.writeTemplateLiteral("#")
		s.cursor.advance(1)
		return nil

	case '\\':
		s.output.WriteRune('\\')
		s.cursor.advance(1)
		return nil

	case '<':
		if !s.cursor.startsWith("<#") {
			s.output.WriteRune('\\')
			return nil
		}

		return s.parseEscapedSpecialTag()

	default:
		s.output.WriteRune('\\')
		return nil
	}
}

func (s *parseSession) writeTemplateLiteral(value string) {
	s.output.WriteString("{{ ")
	s.output.WriteString(strconv.Quote(value))
	s.output.WriteString(" }}")
}

func (s *parseSession) parseEscapedSpecialTag() *diagnostic.Diagnostic {
	const suffix = "#>"

	tagStartIndex := s.cursor.runeIndex()

	for !s.cursor.eof() && !s.cursor.startsWith(suffix) {
		s.cursor.advance(1)
	}

	if s.cursor.eof() {
		literal := s.cursor.slice(
			tagStartIndex,
			s.cursor.runeIndex(),
		)

		s.writeTemplateLiteral(literal)
		return nil
	}

	s.cursor.advance(len(suffix))

	literal := s.cursor.slice(
		tagStartIndex,
		s.cursor.runeIndex(),
	)

	s.writeTemplateLiteral(literal)

	return nil
}
