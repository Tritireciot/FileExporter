package template_compiler

import (
	"fmt"
	"strings"
)

func (parser *Parser) convertIndexes(token string) string {
	for {
		startIdx := strings.Index(token, "[")
		if startIdx == -1 {
			break
		}
		endIdx := strings.Index(token, "]")
		if endIdx == -1 || endIdx < startIdx {
			break
		}

		arrayPart := token[:startIdx]
		indexPart := token[startIdx+1 : endIdx]
		tailPart := token[endIdx+1:]


		arrayPart = parser.ReplaceVar(arrayPart)
		
		if len(arrayPart) > 0 && arrayPart[0] != '$' && arrayPart != "." {
			if len(parser.BlockStack) > 0 {
				var parentVar string
				for i := len(parser.BlockStack) - 1; i >= 0; i-- {
					if parser.BlockStack[i].BlockType != "if" {
						parentVar = parser.BlockStack[i].VarName
						break
					}
				}
				arrayPart = parentVar + "." + arrayPart
			} else {
				arrayPart = "." + arrayPart
			}
		}

		token = fmt.Sprintf("(index %s %s)%s", arrayPart, indexPart, tailPart)
	}
	return token
}