package template_compiler

import (
	"PrintServer/PrintServer/diagnostic"
	"strings"
)

func (p *parser) Parse(source string) ParseResult {
	session := parseSession{
		cursor: cursor{
			source: []rune(source),
			line:   1,
			column: 1,
		},
	}

	return session.parse()
}

func (s *parseSession) parse() ParseResult {
	for !s.cursor.eof() {
		previousIndex := s.cursor.runeIndex()

		if err := s.parseNext(); err != nil {
			return ParseResult{
				Output:      s.output.String(),
				Diagnostics: []diagnostic.Diagnostic{*err},
			}
		}

		if s.cursor.runeIndex() <= previousIndex {
			panic("parseNext did not advance cursor")
		}
	}

	if len(s.blockStack) > 0 {
		scope := s.blockStack[len(s.blockStack)-1]

		message := "Блочная конструкция не закрыта"
		hint := "Добавьте <#End#>"

		switch scope.BlockType {
		case blockTypeRepeat:
			message = "Конструкция Repeat не закрыта"
			hint = "Добавьте <#End#> для закрытия Repeat"

		case blockTypeIf:
			message = "Конструкция If не закрыта"
			hint = "Добавьте <#End#> для закрытия If"

		case blockTypeWith:
			message = "Конструкция With не закрыта"
			hint = "Добавьте <#End#> для закрытия With"
		}

		return ParseResult{
			Output: s.output.String(),
			Diagnostics: []diagnostic.Diagnostic{
				{
					Code:    diagnostic.CodeUnclosedBlock,
					Stage:   diagnostic.StageTranspile,
					Message: message,
					Span: &diagnostic.Span{
						Start: scope.OpenedAt,
						End:   s.cursor.position(),
					},
					Hint: hint,
				},
			},
		}
	}

	return ParseResult{
		Output: s.output.String(),
	}
}

func (s *parseSession) parseNext() *diagnostic.Diagnostic {
	current, ok := s.cursor.current()
	if !ok {
		return nil
	}

	switch {
	case s.cursor.startsWith("{{"):

	case s.cursor.startsWith("}}"):

	case s.cursor.startsWith("\\"):
		return s.parseEscape()

	case s.cursor.startsWith("<#BreakPage#>"):
		s.output.WriteString(`<div style="page-break-before: always;"></div>`)
		s.output.WriteRune('\n')
		s.cursor.advance(len("<#BreakPage#>"))
		return nil

	case s.cursor.startsWith("<#Repeat#"):
		return s.parseRepeat()

	case s.cursor.startsWith("<#With#"):
		return s.parseWith()

	case s.cursor.startsWith("<#If "):
		return s.parseIf()

	case s.cursor.startsWith("<#ElseIf "):
		return s.parseElseIf()

	case s.cursor.startsWith("<#Else#>"):
		return s.parseElse()

	case s.cursor.startsWith("<#End#>"):
		return s.parseEnd()

	case s.cursor.startsWith("<#"):
		return s.parseUnknownSpecialTag()

	case s.cursor.startsWith("#"):
		matched, err := s.tryParseField()
		if matched || err != nil {
			return err
		}
	}

	s.output.WriteRune(current)
	s.cursor.advance(1)

	return nil
}

func (s *parseSession) resolveScopeReference(
	value string,
) string {
	for i := len(s.blockStack) - 1; i >= 0; i-- {
		scope := s.blockStack[i]

		if scope.UserIndexName != "" &&
			value == scope.UserIndexName {
			return scope.IndexName
		}

		if internalName, ok := scope.Fields[value]; ok {
			return internalName
		}

		bestUserName := ""
		bestInternalName := ""

		for userName, internalName := range scope.Fields {
			prefix := userName + "."

			if strings.HasPrefix(value, prefix) &&
				len(userName) > len(bestUserName) {
				bestUserName = userName
				bestInternalName = internalName
			}
		}

		if bestUserName != "" {
			return bestInternalName +
				value[len(bestUserName):]
		}
	}

	return value
}
