package parser

import (
	"PrintServer/html2docx/css"
	"regexp"
	"strings"
)

func getModifier(selector string) (string, string) {
	var modifier string
	if strings.HasSuffix(selector, ":last-child") {
		selector = strings.TrimSuffix(selector, ":last-child")
		modifier = "last-child"
	} else if strings.HasSuffix(selector, ":first-child") {
		selector = strings.TrimSuffix(selector, ":first-child")
		modifier = "first-child"
	}
	return modifier, selector
}

func tokenize(selector string) *css.StyleNode {
	re := regexp.MustCompile(`[>+~]|[^\s>+~]+`)

	tokens := re.FindAllString(selector, -1)
	combinators := map[string]int{
		">": 1,
		"+": 2,
		"~": 3,
	}

	currentIndex := len(tokens) - 1
	modifier, clear_selector := getModifier(tokens[currentIndex])
	styleNode := &css.StyleNode{
		Selector:     clear_selector,
		StyleContent: css.StyleMap{Styles: make(map[css.StyleProperty]css.StyleValue), Order: make([]css.StyleProperty, 0)},
		Modifier:     modifier,
	}

	currentNode := styleNode

	for currentIndex > 0 {
		prevSelector := tokens[currentIndex-1]

		if _, ok := combinators[prevSelector]; ok {
			currentNode.Connection = prevSelector
			modifier, clear_selector := getModifier(tokens[currentIndex-2])
			currentNode.Parent = &css.StyleNode{Selector: clear_selector, Modifier: modifier}
			currentNode = currentNode.Parent
			currentIndex -= 2
		} else {
			currentNode.Connection = " "
			modifier, clear_selector := getModifier(prevSelector)
			currentNode.Parent = &css.StyleNode{Selector: clear_selector, Modifier: modifier}
			currentNode = currentNode.Parent
			currentIndex -= 1

		}
	}

	return styleNode

}

func parseRuleset(templateStyles *css.GlobalStyles, props map[string]string, selectors []string) {
	for _, selectorStr := range selectors {
		selectorStr = strings.TrimSpace(selectorStr)
		complexSelector := tokenize(selectorStr)
		for property, value := range props {
			if complexSelector.Parent == nil && complexSelector.Modifier == "" {
				if strings.HasPrefix(selectorStr, ".") {
					css.InsertSelectorStyles(templateStyles.Classes.ClassStyles, strings.TrimPrefix(selectorStr, "."), property, value)
					templateStyles.Classes.Order = append(templateStyles.Classes.Order, strings.TrimPrefix(selectorStr, "."))
				} else if strings.HasPrefix(selectorStr, "#") {
					css.InsertSelectorStyles(templateStyles.IDs, strings.TrimPrefix(selectorStr, "#"), property, value)
				} else if selectorStr == "*" || selectorStr == "body" {
					templateStyles.DefaultStyle.Order = append(templateStyles.DefaultStyle.Order, css.StyleProperty(property))
					templateStyles.DefaultStyle.Styles[css.StyleProperty(property)] = css.StyleValue(value)
				} else {
					css.InsertSelectorStyles(templateStyles.Tags, selectorStr, property, value)
				}
			} else {
				complexSelector.StyleContent.Order = append(complexSelector.StyleContent.Order, css.StyleProperty(property))
				complexSelector.StyleContent.Styles[css.StyleProperty(property)] = css.StyleValue(value)
			}
		}
		if complexSelector.Parent != nil || complexSelector.Modifier != "" {
			templateStyles.ComplexStyles = append(templateStyles.ComplexStyles, complexSelector)
		}
	}
}
