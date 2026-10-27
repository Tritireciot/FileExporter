package html

import (
	nethtml "golang.org/x/net/html"
)

func FindTag(node *nethtml.Node, tagName Tag) []*nethtml.Node {

	if node == nil {
		return nil
	}

	var nodes []*nethtml.Node
	var traverse func(node *nethtml.Node)
	traverse = func(currentNode *nethtml.Node) {
		if currentNode.Type == nethtml.ElementNode && currentNode.Data == string(tagName) {
			nodes = append(nodes, currentNode)
		}
		for child := currentNode.FirstChild; child != nil; child = child.NextSibling {
			traverse(child)
		}
	}
	traverse(node)

	return nodes
}

func IterateOverTags(nodes []*nethtml.Node, iterateFunc func(node *nethtml.Node)) {
	for _, node := range nodes {
		if node.FirstChild != nil && node.FirstChild.Type == nethtml.TextNode {
			iterateFunc(node)
		}
	}
}

func IterateOverChildren(parentNode *nethtml.Node, iterateFunc func(node *nethtml.Node)) {
	for child := parentNode.FirstChild; child != nil; child = child.NextSibling {
		iterateFunc(child)
	}
}

func IterateOverSiblings(node *nethtml.Node, iterateFunc func(node *nethtml.Node)) {
	for next_node := node; next_node != nil; next_node = next_node.NextSibling {
		iterateFunc(next_node)
	}

}

func GetAttribute(node *nethtml.Node, attrName HTMLAttr) (string, bool) {
	for _, attr := range node.Attr {
		if attr.Key == string(attrName) {
			return attr.Val, true
		}
	}
	return "", false
}
