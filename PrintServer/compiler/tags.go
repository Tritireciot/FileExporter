package template_compiler

import (
	"fmt"
	"strings"
)

func cleanVarName(path string) string {
	if idx := strings.LastIndex(path, "."); idx != -1 {
		path = path[idx+1:]
	}
	if idx := strings.Index(path, "["); idx != -1 {
		path = path[:idx]
	}
	return path
}

func (parser *Parser) buildContextPath(newFieldName string) string {
	if strings.Contains(newFieldName, "[") && strings.Contains(newFieldName, "]") {
		return parser.convertIndexes(newFieldName)
	}
	if len(parser.BlockStack) == 0 {
		return "." + newFieldName
	}

	var parentVar string
	for i := len(parser.BlockStack) - 1; i >= 0; i-- {
		if parser.BlockStack[i].BlockType != "if" {
			parentVar = parser.BlockStack[i].VarName
			break
		}
	}
	return parentVar + "." + newFieldName
}

func (parser *Parser) parseSpecialTags(candidateSyntax string) (bool, string) {
	if candidateSyntax == "#End#" {
		if len(parser.BlockStack) == 0 {
			return true, "{{ end }}"
		}
		parser.BlockStack[len(parser.BlockStack)-1] = Scope{}
		parser.BlockStack = parser.BlockStack[:len(parser.BlockStack)-1]
		return true, "{{ end }}"
	}

	if candidateSyntax == "#BreakPage#" {
		return true, "<div style=\"page-break-before: always;\"></div>\n"
	}

	if strings.HasPrefix(candidateSyntax, "#Repeat#") {
		rawFieldName := candidateSyntax[len("#Repeat#") : len(candidateSyntax)-1]
		cleanName := cleanVarName(rawFieldName)

		parser.Encounter++
		contextVar := fmt.Sprintf("$%s%d", cleanName, parser.Encounter)		

		parser.Encounter++
		contextIndex := fmt.Sprintf("$%sIndex%d", cleanName, parser.Encounter) 

		userIndexName := cleanName + "Index"
		
		rangePath := parser.buildContextPath(rawFieldName)

		parser.BlockStack = append(parser.BlockStack, Scope{
			BlockType: "range",
			VarName:   contextVar,
			IndexName: contextIndex, 
			UserIndexName: userIndexName,
			Fields:    map[string]string{cleanName: contextVar},
		})

		return true, fmt.Sprintf("{{ range %s, %s := %s }}", contextIndex, contextVar, rangePath)
	}


	if strings.HasPrefix(candidateSyntax, "#If ") {
		condition := candidateSyntax[len("#If ") : len(candidateSyntax)-1]
		parser.BlockStack = append(parser.BlockStack, Scope{
			BlockType: "if",
		})
		return true, fmt.Sprintf("{{ if %s }}", parser.parseCondition(condition))
	}

	if strings.HasPrefix(candidateSyntax, "#ElseIf ") {
		condition := candidateSyntax[len("#ElseIf ") : len(candidateSyntax)-1]
		return true, fmt.Sprintf("{{ else if %s }}", parser.parseCondition(condition))
	}

	if candidateSyntax == "#Else#" {
		return true, "{{ else }}"
	}

	if strings.HasPrefix(candidateSyntax, "#With#") {
		parts := strings.Split(candidateSyntax[len("#With#"):len(candidateSyntax)-1], "#")
		if len(parts) != 2 { return false, "" }
		
		varName, fieldName := parts[0], parts[1]
		withPath := parser.buildContextPath(fieldName)
		contextVar := "$" + varName

		// Создаем изолированный Scope
		parser.BlockStack = append(parser.BlockStack, Scope{
			BlockType: "with",
			VarName:   contextVar,
			Fields:    map[string]string{varName: contextVar},
		})

		return true, fmt.Sprintf("{{ with %s := %s }}", contextVar, withPath)
}

	return false, ""
}
