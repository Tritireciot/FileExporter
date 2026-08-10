package parser

import (
	"PrintServer/html2docx/converters"
	"PrintServer/html2docx/css"
	"PrintServer/html2docx/html"

	"github.com/mmonterroca/docxgo/v2/domain"
	nethtml "golang.org/x/net/html"
)

func (p *DocumentParser) handleDefaultBlock(node *nethtml.Node, tagName html.Tag, state TagStyleState, tagStyle css.StyleMap, source_doc Source) {
	paragraph, _ := source_doc.AddParagraph()

	switch tagName {
	case html.LinkTag:
		if href, ok := html.GetAttribute(node, html.HREFAttr); ok {
			state.Link = href
			tagStyle.Order = append(tagStyle.Order, css.Color)
			tagStyle.Styles[css.Color] = css.CssBlue
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
			if processedImage, err := converters.ConvertImage(source); err == nil {
				paragraph.AddImageFromBytesWithSize(processedImage.Data, processedImage.Format, processedImage.Size)
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
