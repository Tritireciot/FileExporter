package parser

import (
	"PrintServer/html2docx/css"
	"PrintServer/html2docx/html"
	"PrintServer/html2docx/utils"

	"github.com/mmonterroca/docxgo/v2/domain"
	nethtml "golang.org/x/net/html"
)

type DocumentParser struct {
	docs           domain.Document
	templateStyles css.GlobalStyles
}

func NewDocumentParser(docs domain.Document, styles css.GlobalStyles) *DocumentParser {
	return &DocumentParser{docs: docs, templateStyles: styles}
}

func (p *DocumentParser) Convert(rootNode *nethtml.Node) {
	section, _ := p.docs.DefaultSection()

	section.SetPageSize(p.templateStyles.Page.Size)
	section.SetOrientation(p.templateStyles.Page.Orientation)
	section.SetMargins(p.templateStyles.Page.Margins)

	bodyTags := html.FindTag(rootNode, html.BodyTag)
	html.IterateOverChildren(bodyTags[0], func(node *nethtml.Node) {
		if node.Type == nethtml.ElementNode {
			p.parseNode(node, p.templateStyles.DefaultStyle)
		}
	})
}

func (p *DocumentParser) parseNode(node *nethtml.Node, currentStyles css.StyleMap) {
	tagName := html.Tag(node.Data)

	if tagName == html.StyleTag || node.Type != nethtml.ElementNode {
		return
	}
	if html.IsPageBreak(node) {
		p.docs.AddPageBreak()
		return
	}

	tagStyle := utils.CopyMap(currentStyles)
	p.templateStyles.CombineStyles(node, tagStyle)

	state := TagStyleState{}

	switch {
	case tagName.IsContainer():
		html.IterateOverChildren(node, func(child *nethtml.Node) {
			p.parseNode(child, tagStyle)
		})

	case tagName.IsList():
		p.handleList(node, tagName, state, tagStyle)

	case tagName == html.TableTag:
		p.handleTable(node, tagStyle)

	default:

		p.handleDefaultBlock(node, tagName, state, tagStyle)
	}
}
