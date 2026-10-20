package template_compiler

import (
	"fmt"
	"strings"
)

func (parser *Parser) Parse(templateStr string) string {
	defer func() {
		parser.BlockStack = nil 
		parser.Encounter = 0 
	}()

	runes := []rune(templateStr)
	var result strings.Builder
	i := 0

	for i < len(runes) {
		if i+1 < len(runes) {
			if runes[i] == '{' && runes[i+1] == '{' {
				result.WriteString(`{{ "{{" }}`)
				i += 2
				continue
			}
			if runes[i] == '}' && runes[i+1] == '}' {
				result.WriteString(`{{ "}}" }}`)
				i += 2
				continue
			}

			if runes[i] == '\\' {
				if ok, change, add := checkEscapeCharacters(runes, i+1); ok {
					result.WriteString(change)
					i += add
					continue
				}
			}

			if runes[i] == '<' && runes[i+1] == '#' {
				endIndex := i + 2
				closed := false
				isValid := true

				for endIndex < len(runes) {
					if runes[endIndex] == '#' && endIndex+1 < len(runes) && runes[endIndex+1] == '>' {
						closed = true
						endIndex++
						break
					}
					if !isValidFieldChar(runes[endIndex]) {
						isValid = false
						break
					}
					endIndex++
				}

				if closed && isValid {
					if ok, parsed := parser.parseSpecialTags(string(runes[i+1 : endIndex])); ok {
						result.WriteString(parsed)
						i = endIndex + 1
						continue
					}
				}
			}

			if runes[i] == '#' {
				endIndex := i + 1
				closed := false
				isValid := true

				for endIndex < len(runes) {
					if runes[endIndex] == '#' {
						closed = true
						break
					}
					if !isValidFieldChar(runes[endIndex]) {
						isValid = false
						break
					}
					endIndex++
				}

				if closed && isValid && endIndex-i > 2 {
					rawContent := string(runes[i+1 : endIndex])
					
					var tokens []string
					var current strings.Builder
					
					for _, r := range rawContent {
						if r == ' ' || r == '|' {
							if current.Len() > 0 {
								tokens = append(tokens, current.String())
								current.Reset()
							}
							if r == '|' {
								tokens = append(tokens, "|")
							}
						} else {
							current.WriteRune(r)
						}
					}
					if current.Len() > 0 {
						tokens = append(tokens, current.String())
					}

					for idx, token := range tokens {
						if token == "|" {
							continue
						}

						isFunc := idx > 0 && tokens[idx-1] == "|"

						if strings.Contains(token, "[") && strings.Contains(token, "]") {
							tokens[idx] = parser.convertIndexes(token)
						} else {
							replaced := parser.ReplaceVar(token)

							if strings.HasPrefix(replaced, "$") {
								tokens[idx] = replaced
							} else if isFunc || strings.HasPrefix(replaced, `"`) || strings.HasPrefix(replaced, `'`) || replaced == "." {
								tokens[idx] = replaced
							} else {
								tokens[idx] = "." + replaced
							}
						}
					}

					var finalContent strings.Builder
					for idx, token := range tokens {
						if idx > 0 {
							finalContent.WriteString(" ")
						}
						finalContent.WriteString(token)
					}

					result.WriteString(fmt.Sprintf("{{ %s }}", finalContent.String()))
					i = endIndex + 1
					continue
				}
			}
		}

		result.WriteRune(runes[i])
		i++
	}
	return result.String()
}