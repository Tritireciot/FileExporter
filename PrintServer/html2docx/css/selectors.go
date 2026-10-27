package css

import (
	"PrintServer/PrintServer/html2docx/html"
	"strings"

	nethtml "golang.org/x/net/html"
)

func getClass(node *nethtml.Node) *string {
	for _, attribute := range node.Attr {
		if html.HTMLAttr(attribute.Key) == html.ClassAttr {
			return &attribute.Val
		}
	}
	return nil
}

func getID(node *nethtml.Node) *string {
	for _, attribute := range node.Attr {
		if html.HTMLAttr(attribute.Key) == html.IdAttr {
			return &attribute.Val
		}
	}
	return nil
}

type SelectorChecker func(node *nethtml.Node, styleNode *StyleNode) (bool, *nethtml.Node)

var SelectorCheckers = map[string]SelectorChecker{
	" ": func(node *nethtml.Node, styleNode *StyleNode) (bool, *nethtml.Node) {

		currentNode := node.Parent
		for currentNode.Parent != nil {
			if currentNode.Type == nethtml.ElementNode && CheckSelector(currentNode, styleNode.Parent) {
				return true, currentNode
			}
			currentNode = currentNode.Parent
		}

		return false, node
	},
	">": func(node *nethtml.Node, styleNode *StyleNode) (bool, *nethtml.Node) {
		if node.Parent != nil && node.Parent.Type == nethtml.ElementNode && CheckSelector(node.Parent, styleNode.Parent) {
			return true, node.Parent
		}
		return false, node
	},
	"+": func(node *nethtml.Node, styleNode *StyleNode) (bool, *nethtml.Node) {
		currentNode := node.PrevSibling
		for currentNode != nil && currentNode.Type != nethtml.ElementNode {
			currentNode = currentNode.PrevSibling
		}
		if currentNode != nil && CheckSelector(currentNode, styleNode.Parent) {
			return true, currentNode
		}
		return false, node
	},
	"~": func(node *nethtml.Node, styleNode *StyleNode) (bool, *nethtml.Node) {
		currentNode := node.PrevSibling
		for currentNode != nil {
			if currentNode.Type == nethtml.ElementNode && CheckSelector(currentNode, styleNode.Parent) {
				return true, currentNode
			}
			currentNode = currentNode.PrevSibling
		}
		return false, node
	},
	"last-child": func(node *nethtml.Node, styleNode *StyleNode) (bool, *nethtml.Node) {
		currentNode := node.NextSibling
		for currentNode != nil {
			if currentNode.Type == nethtml.ElementNode {
				return false, node
			}
			currentNode = currentNode.NextSibling
		}
		return true, node
	},
	"first-child": func(node *nethtml.Node, styleNode *StyleNode) (bool, *nethtml.Node) {
		currentNode := node.PrevSibling
		for currentNode != nil {
			if currentNode.Type == nethtml.ElementNode {
				return false, node
			}
			currentNode = currentNode.PrevSibling
		}
		return true, node
	},
}

func CheckSelector(node *nethtml.Node, styleNode *StyleNode) bool {
	if node == nil || styleNode == nil {
		return false
	}

	if strings.HasPrefix(styleNode.Selector, ".") {
		nodeClass := getClass(node)
		if nodeClass == nil || *nodeClass != strings.TrimPrefix(styleNode.Selector, ".") {
			return false
		}
	} else if strings.HasPrefix(styleNode.Selector, "#") {
		nodeId := getID(node)
		if nodeId == nil || *nodeId != strings.TrimPrefix(styleNode.Selector, "#") {
			return false
		}
	} else {
		if styleNode.Selector != node.Data {
			return false
		}
	}

	return true
}
