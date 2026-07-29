package css

import (
	"PrintServer/html2docx/html"
	"strings"

	nethtml "golang.org/x/net/html"
)

func (templateStyles GlobalStyles) CombineStyles(node *nethtml.Node, currentStyles StyleMap) {
	tagName := node.Data

	if defaultStyles, ok := DefaultTagStyles[html.Tag(tagName)]; ok {
		for prop, value := range defaultStyles {
			currentStyles[prop] = value
		}
	}

	for tag, tagStyle := range templateStyles.Tags {
		if tag == tagName {
			for prop, value := range tagStyle {
				currentStyles[prop] = value
			}
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
		for className, classStyle := range templateStyles.Classes {
			if strings.Contains(classStr, className) {
				for prop, value := range classStyle {
					currentStyles[prop] = value
				}
			}
		}
	}

	if tagID != "" {
		if tagStyle, ok := templateStyles.Tags[tagID]; ok {
			for prop, value := range tagStyle {
				currentStyles[prop] = value
			}
		}
	}

	if tagStyle != "" {
		ParseLocalTagStyle(tagStyle, currentStyles)
	}
}
