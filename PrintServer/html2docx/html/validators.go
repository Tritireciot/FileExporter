package html

import (
	"strings"

	nethtml "golang.org/x/net/html"
)

func IsPageBreak(node *nethtml.Node) bool {
	if Tag(node.Data) != DivTag {
		return false
	}

	style, ok := GetAttribute(node, StyleAttr)
	if !ok {
		return false
	}

	return strings.Contains(style, "page-break-before: always;")
}
