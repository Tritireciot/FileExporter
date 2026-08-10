package css

import (
	"PrintServer/html2docx/html"
	"strings"

	nethtml "golang.org/x/net/html"
)

func (templateStyles GlobalStyles) CombineStyles(node *nethtml.Node, currentStyles *StyleMap) {
	tagName := node.Data

	if defaultStyles, ok := DefaultTagStyles[html.Tag(tagName)]; ok {
		currentStyles.Update(defaultStyles)
	}

	for tag, tagStyle := range templateStyles.Tags {
		if tag == tagName {
			currentStyles.Update(tagStyle)
		}
	}
	classStr := ""
	tagID := ""
	tagStyle := ""

	for _, attribute := range node.Attr {
		switch html.HTMLAttr(attribute.Key) {
		case html.ClassAttr:
			classStr = attribute.Val
		case html.IdAttr:
			tagID = attribute.Val
		case html.StyleAttr:
			tagStyle = attribute.Val
		}
	}
	if classStr != "" {
		for _, className := range templateStyles.Classes.Order {
			if strings.Contains(classStr, className) {
				currentStyles.Update(templateStyles.Classes.ClassStyles[className])
			}
		}
	}

	if tagID != "" {
		if tagStyle, ok := templateStyles.Tags[tagID]; ok {
			currentStyles.Update(tagStyle)
		}
	}

	if tagStyle != "" {
		ParseLocalTagStyle(tagStyle, currentStyles)
	}
}
