package parser

import (
	"PrintServer/html2docx/css"
	"PrintServer/html2docx/html"
	"PrintServer/html2docx/utils"
	"strings"

	docx "github.com/mmonterroca/docxgo/v2"
	"github.com/mmonterroca/docxgo/v2/domain"
	nethtml "golang.org/x/net/html"
)

func (p *DocumentParser) parseInlineElements(paragraph domain.Paragraph, node *nethtml.Node, state TagStyleState, currentRawStyles css.StyleMap) {
	if node.Type == nethtml.TextNode && len(strings.TrimSpace(node.Data)) > 0 {
		run, _ := paragraph.AddRun()

		styleModel := BuildStyleModel(currentRawStyles)
		styleModel.TextPrefix = state.TextPrefix

		styleModel.ApplyToRun(&run)

		finalText := styleModel.TransformText(node.Data)

		if state.Link != "" {
			linkField := docx.NewHyperlinkField(state.Link, finalText)
			run.AddField(linkField)
		} else {
			run.SetText(finalText)
		}
		return
	}

	if node.Type == nethtml.ElementNode {
		childRawStyles := utils.CopyMap(currentRawStyles)

		p.templateStyles.CombineStyles(node, childRawStyles)

		switch html.Tag(node.Data) {
		case html.LinkTag:
			if href, ok := html.GetAttribute(node, html.HREFAttr); ok {
				state.Link = href
				childRawStyles["color"] = "blue"
			}
		case html.BrTag:
			run, _ := paragraph.AddRun()
			run.AddBreak(domain.BreakTypeLine)
		}

		for child := node.FirstChild; child != nil; child = child.NextSibling {
			p.parseInlineElements(paragraph, child, state, childRawStyles)
		}
	}
}
