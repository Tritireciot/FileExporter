package css

import (
	"strings"

	"github.com/tdewolff/parse/v2"
	parceCSS "github.com/tdewolff/parse/v2/css"
)

func ParseLocalTagStyle(parceCSS_string string, tagStyles StyleMap) {
	parser := parceCSS.NewParser(parse.NewInputString(parceCSS_string), true)
	for {
		grammar_type, _, data := parser.Next()
		if grammar_type == parceCSS.ErrorGrammar {
			break
		}

		switch grammar_type {
		case parceCSS.DeclarationGrammar:
			tagStyles[StyleProperty(string(data))] = StyleValue(TokensToString(parser.Values()))
		}
	}
}

func InsertSelectorStyles(styleMap map[string]StyleMap, selectorStr, property, value string) {
	selectorStyles, ok := styleMap[selectorStr]
	if !ok {
		selectorStyles = StyleMap{}
	}
	selectorStyles[StyleProperty(property)] = StyleValue(value)
	styleMap[selectorStr] = selectorStyles
}

func ParsePropsInto(p *parceCSS.Parser, props map[string]string) {
	for {
		gt, _, data := p.Next()
		if gt == parceCSS.DeclarationGrammar {
			prop := string(data)
			props[prop] = TokensToString(p.Values())

		}
		if gt == parceCSS.EndRulesetGrammar || gt == parceCSS.EndAtRuleGrammar || gt == parceCSS.ErrorGrammar {
			break
		}
	}
}

func TokensToString(tokens []parceCSS.Token) string {
	var sb strings.Builder
	for _, v := range tokens {
		sb.Write(v.Data)
	}
	return strings.TrimSpace(sb.String())
}
