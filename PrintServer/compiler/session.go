package template_compiler

import "strings"

type parseSession struct {
	cursor cursor

	output     strings.Builder
	blockStack []Scope
	encounter  int
}

func isDataScope(scope Scope) bool {
	return scope.BlockType == blockTypeRepeat ||
		scope.BlockType == blockTypeWith
}

func (s *parseSession) buildContextPath(fieldName string) string {

	for i := len(s.blockStack) - 1; i >= 0; i-- {
		scope := s.blockStack[i]

		if !isDataScope(scope) {
			continue
		}

		if internalName, ok := scope.Fields[fieldName]; ok {
			return internalName
		}

		bestUserName := ""
		bestInternalName := ""

		for userName, internalName := range scope.Fields {
			prefix := userName + "."

			if strings.HasPrefix(fieldName, prefix) &&
				len(userName) > len(bestUserName) {
				bestUserName = userName
				bestInternalName = internalName
			}
		}

		if bestUserName != "" {
			return bestInternalName +
				fieldName[len(bestUserName):]
		}
	}

	for i := len(s.blockStack) - 1; i >= 0; i-- {
		scope := s.blockStack[i]

		if !isDataScope(scope) {
			continue
		}

		return scope.VarName + "." + fieldName
	}

	return "." + fieldName
}
