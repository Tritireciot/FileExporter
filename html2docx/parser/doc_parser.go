package parser

import (
	"PrintServer/html2docx/css"
	"PrintServer/html2docx/html"
	"strings"

	"github.com/mmonterroca/docxgo/v2/domain"
	nethtml "golang.org/x/net/html"
)

type Source interface {
	AddTable(rows, cols int) (domain.Table, error)
	AddParagraph() (domain.Paragraph, error)
}

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
			p.parseNode(node, p.templateStyles.DefaultStyle, p.docs)
		}
	})
}

func (p *DocumentParser) parseNode(node *nethtml.Node, currentStyles css.StyleMap, source Source) {

	if node == nil {
		return
	}
	tagName := html.Tag(node.Data)

	if tagName == html.StyleTag {
		return
	}
	if html.IsPageBreak(node) {
		p.docs.AddPageBreak()
		return
	}

	tagStyle := currentStyles.Copy()
	p.templateStyles.CombineStyles(node, &tagStyle)

	state := TagStyleState{}

	switch {
	case tagName.IsContainer():
		html.IterateOverChildren(node, func(child *nethtml.Node) {
			p.parseNode(child, tagStyle, source)
		})

	case tagName.IsList():
		p.handleList(node, tagName, state, tagStyle)

	case tagName == html.TableTag:
		p.handleTable(node, tagStyle, source)
	case node.Type == nethtml.TextNode:
		if len(strings.TrimSpace((node.Data))) > 0 {
			paragraph, _ := source.AddParagraph()
			styleModel := BuildStyleModel(currentStyles)
			styleModel.ApplyToParagraph(&paragraph)
			p.parseInlineElements(paragraph, node, TagStyleState{}, currentStyles)
		}
	default:

		p.handleDefaultBlock(node, tagName, state, tagStyle, source)
	}
}
