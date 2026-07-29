package parser

import (
	"PrintServer/html2docx/css"
	"PrintServer/html2docx/html"
	"PrintServer/html2docx/utils"
	"fmt"

	nethtml "golang.org/x/net/html"
)

func (p *DocumentParser) handleList(node *nethtml.Node, tagName html.Tag, state TagStyleState, tagStyle css.StyleMap) {

	if tagName == html.OLTag {
		state.Order = 0
	} else {
		state.TextPrefix = "• "
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type != nethtml.ElementNode {
			continue
		}

		paragraph, _ := p.docs.AddParagraph()

		if tagName == html.OLTag {
			state.Order += 1
			state.TextPrefix = fmt.Sprintf("%d. ", state.Order)
		}

		childStyle := utils.CopyMap(tagStyle)
		p.templateStyles.CombineStyles(child, childStyle)

		styleModel := BuildStyleModel(childStyle)

		styleModel.TextPrefix = state.TextPrefix

		styleModel.ApplyToParagraph(&paragraph)

		for liChild := child.FirstChild; liChild != nil; liChild = liChild.NextSibling {
			p.parseInlineElements(paragraph, liChild, state, childStyle)
		}
	}
}
