package parser

import (
	"PrintServer/html2docx/css"
	"PrintServer/html2docx/html"
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

func NewTemplateStyles(rootNode *nethtml.Node) *css.GlobalStyles {
	templateStyles := css.GlobalStyles{
		Classes: make(map[string]css.StyleMap),
		IDs:     make(map[string]css.StyleMap),
		Tags:    make(map[string]css.StyleMap),
	}

	templateStyles.DefaultStyle = css.StyleMap{}

	templateStyles.Page.Size = domain.PageSizeA4
	templateStyles.Page.Orientation = domain.OrientationPortrait
	templateStyles.Page.Margins = domain.DefaultMargins

	styleTags := html.FindTag(rootNode, html.StyleTag)
	html.IterateOverTags(styleTags, func(node *nethtml.Node) {
		CollectGlobalStyles(strings.TrimSpace(node.FirstChild.Data), &templateStyles)
	})

	return &templateStyles
}
