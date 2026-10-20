package template_compiler

import "fmt"

func checkEscapeCharacters(runes []rune, index int) (bool, string, int) {
	if len(runes) <= index {
		return false, "", 0
	}

	switch runes[index] {
	case '#':
		return true, `{{ "#" }}`, 2
	case '\\':
		return true, `\`, 2
	case '<':
		if index+1 < len(runes) && runes[index+1] == '#' {
			endIndex := index + 1
			closed := false
			for endIndex < len(runes) {
				if runes[endIndex] == '#' && endIndex+1 < len(runes) && runes[endIndex+1] == '>' {
					closed = true
					endIndex++
					break
				}
				endIndex++
			}
			if closed {
				res := string(runes[index : endIndex+1])
				return true, fmt.Sprintf(`{{ "%s" }}`, res), len(res) + 1
			}
			return true, `{{ "<#" }}`, 3
		}
	}
	return false, "", 0
}
