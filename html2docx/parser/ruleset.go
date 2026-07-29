package parser

import (
	"PrintServer/html2docx/css"
	"strings"
)

func parseRuleset(templateStyles *css.GlobalStyles, props map[string]string, selectors []string) {

	for property, value := range props {
		for _, selectorStr := range selectors {
			selectorStr = strings.Trim(selectorStr, " ")
			if strings.HasPrefix(selectorStr, ".") {
				selectorStr = strings.TrimPrefix(selectorStr, ".")
				css.InsertSelectorStyles(templateStyles.Classes, selectorStr, property, value)
			} else if strings.HasPrefix(selectorStr, "#") {
				selectorStr = strings.TrimPrefix(selectorStr, "#")
				css.InsertSelectorStyles(templateStyles.IDs, selectorStr, property, value)
			} else if selectorStr == "*" || selectorStr == "body" {
				templateStyles.DefaultStyle[css.StyleProperty(property)] = css.StyleValue(value)
			} else {
				css.InsertSelectorStyles(templateStyles.Tags, selectorStr, property, value)
			}
		}
	}
}
