package parser

import (
	"PrintServer/PrintServer/html2docx/css"
	"PrintServer/PrintServer/html2docx/html"
	"fmt"
	"strings"

	"github.com/mmonterroca/docxgo/v2/domain"
	"github.com/tdewolff/parse/v2"
	parseCSS "github.com/tdewolff/parse/v2/css"
	nethtml "golang.org/x/net/html"
)

func CollectGlobalStyles(css_string string, templateStyles *css.GlobalStyles) {
	parser := parseCSS.NewParser(parse.NewInputString(css_string), false)
	for {
		grammar_type, _, data := parser.Next()
		if grammar_type == parseCSS.ErrorGrammar {
			break
		}
		props := make(map[string]string)
		switch grammar_type {
		case parseCSS.BeginAtRuleGrammar:
			atRuleName := css.StyleProperty(string(data))
			for {
				nextGt, _, nextData := parser.Next()
				if nextGt == parseCSS.DeclarationGrammar || nextGt == parseCSS.EndAtRuleGrammar {

					prop := string(nextData)
					props[prop] = css.TokensToString(parser.Values())

					css.ParsePropsInto(parser, props)

					if atRuleName == css.Page {
						parsePageRule(templateStyles, props)
					}
					break
				}
			}
		case parseCSS.BeginRulesetGrammar:
			selectorsStr := css.TokensToString(parser.Values())
			selectors := strings.Split(selectorsStr, ",")
			css.ParsePropsInto(parser, props)
			parseRuleset(templateStyles, props, selectors)

		}

	}
}

func checkNode(node *nethtml.Node, styleNode *css.StyleNode) bool {
	if node.Type != nethtml.ElementNode {
		return false
	}

	// Самая первая проверка: совпадает ли вообще базовый целевой тег/класс
	if !css.CheckSelector(node, styleNode) {
		return false
	}

	currentStyleNode := styleNode
	currentNode := node
	for currentStyleNode != nil {
		if currentStyleNode.Modifier != "" {
			checker, ok := css.SelectorCheckers[currentStyleNode.Modifier]
			if !ok {
				return false
			}
			ok, _ = checker(currentNode, currentStyleNode)
			if !ok {
				return false
			}
		}

		if currentStyleNode.Parent == nil {
			break
		}
		checker, ok := css.SelectorCheckers[currentStyleNode.Connection]
		if !ok {
			return false
		}

		ok, currentNode = checker(currentNode, currentStyleNode)
		if !ok {
			return false
		}

		currentStyleNode = currentStyleNode.Parent
	}

	return true
}

func modifyNode(node *nethtml.Node, ComplexStyles []*css.StyleNode) {
	for _, styleNode := range ComplexStyles {
		if checkNode(node, styleNode) {
			var styleStr string
			var attrIndex int = -1

			for i, attribute := range node.Attr {
				if html.HTMLAttr(attribute.Key) == html.StyleAttr {
					styleStr = attribute.Val
					attrIndex = i
					break
				}
			}

			styleStr += styleNode.StyleContent.String()
			if attrIndex != -1 {
				node.Attr[attrIndex].Val = styleStr
				fmt.Println(node.Attr[attrIndex].Namespace)
			} else if styleStr != "" {
				node.Attr = append(node.Attr, nethtml.Attribute{Key: string(html.StyleAttr), Val: styleStr})
			}

		}
	}

	html.IterateOverChildren(node, func(child *nethtml.Node) {
		modifyNode(child, ComplexStyles)
	})

}

func preprocessComplexStyles(rootNode *nethtml.Node, ComplexStyles []*css.StyleNode) {

	bodyTag := html.FindTag(rootNode, html.BodyTag)
	html.IterateOverChildren(bodyTag[0], func(node *nethtml.Node) {
		modifyNode(node, ComplexStyles)
	})

}

func NewTemplateStyles(rootNode *nethtml.Node) *css.GlobalStyles {
	templateStyles := css.GlobalStyles{
		Classes: css.ClassRule{
			Order:       make([]string, 0),
			ClassStyles: make(map[string]css.StyleMap),
		},
		IDs:           make(map[string]css.StyleMap),
		Tags:          make(map[string]css.StyleMap),
		ComplexStyles: make([]*css.StyleNode, 0),
	}

	templateStyles.DefaultStyle = css.StyleMap{
		Styles: make(map[css.StyleProperty]css.StyleValue),
		Order:  make([]css.StyleProperty, 0),
	}

	templateStyles.Page.Size = domain.PageSizeA4
	templateStyles.Page.Orientation = domain.OrientationPortrait
	templateStyles.Page.Margins = domain.DefaultMargins

	styleTags := html.FindTag(rootNode, html.StyleTag)
	html.IterateOverTags(styleTags, func(node *nethtml.Node) {
		CollectGlobalStyles(strings.TrimSpace(node.FirstChild.Data), &templateStyles)
	})

	preprocessComplexStyles(rootNode, templateStyles.ComplexStyles)

	return &templateStyles
}
