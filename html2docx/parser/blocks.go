package parser

import (
	"PrintServer/html2docx/css"
	"PrintServer/html2docx/html"
	"encoding/base64"
	"strings"

	"github.com/mmonterroca/docxgo/v2/domain"
	nethtml "golang.org/x/net/html"
)

func (p *DocumentParser) handleDefaultBlock(node *nethtml.Node, tagName html.Tag, state TagStyleState, tagStyle css.StyleMap) {
	paragraph, _ := p.docs.AddParagraph()

	switch tagName {
	case html.LinkTag:
		if href, ok := html.GetAttribute(node, html.HREFAttr); ok {
			state.Link = href
			tagStyle["color"] = "blue"
		}
	case html.DDTag:
		state.TextPrefix = "        "
	case html.BlockQuoteTag:
		paragraph.SetStyle(domain.StyleIDIntenseQuote)
	case html.HrTag:
		paragraph.SetBorderBottom(domain.BorderStyle{
			Style: domain.BorderSingle,
			Width: 1,
			Color: domain.ColorBlack,
		})
		return
	case html.ImgTag:
		if source, ok := html.GetAttribute(node, html.SourceAttr); ok {
			if strings.HasPrefix(source, "data:image/png;base64,") {
				imgBytes, _ := base64.StdEncoding.DecodeString(strings.TrimPrefix(source, "data:image/png;base64,"))
				paragraph.AddImageFromBytes(imgBytes, domain.ImageFormatPNG)
			}
		}
		return
	}

	styleModel := BuildStyleModel(tagStyle)
	styleModel.TextPrefix = state.TextPrefix

	styleModel.ApplyToParagraph(&paragraph)

	html.IterateOverChildren(node, func(child *nethtml.Node) {
		p.parseInlineElements(paragraph, child, state, tagStyle)
	})
}
