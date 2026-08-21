package parser

import (
	"PrintServer/PrintServer/html2docx/css"
	"PrintServer/PrintServer/html2docx/html"

	nethtml "golang.org/x/net/html"
)

func (p *DocumentParser) handleList(node *nethtml.Node, tagName html.Tag, state TagStyleState, tagStyle css.StyleMap) {

	state.ListLevel += 1

	wordLevel := state.ListLevel - 1

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type != nethtml.ElementNode {
			continue
		}

		paragraph, _ := p.docs.AddParagraph()

		childStyle := tagStyle.Copy()
		p.templateStyles.CombineStyles(child, &childStyle)

		if _, hasType := childStyle.Styles[css.ListStyleType]; !hasType {
			childStyle.Order = append(childStyle.Order, css.ListStyleType)
			if tagName == html.OLTag {
				childStyle.Styles[css.ListStyleType] = css.ListStyleDecimal
			} else {
				childStyle.Styles[css.ListStyleType] = css.ListStyleDisc
			}
		}

		styleModel := BuildStyleModel(childStyle)
		styleModel.ApplyToParagraph(&paragraph)

		if ref, ok := paragraph.Numbering(); ok {
			ref.Level = wordLevel
			paragraph.SetNumbering(ref)
		}

		for liChild := child.FirstChild; liChild != nil; liChild = liChild.NextSibling {
			p.parseInlineElements(paragraph, liChild, state, childStyle)
		}
	}
}
