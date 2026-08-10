package parser

import (
	"PrintServer/html2docx/converters"
	"PrintServer/html2docx/css"

	"github.com/mmonterroca/docxgo/v2/domain"
)

type PropertyParser func(value css.StyleValue, model *StyleModel)

var parsers = map[css.StyleProperty]PropertyParser{

	css.Color: func(value css.StyleValue, model *StyleModel) {
		if color := converters.ConvertColor(string(value)); color != nil {
			model.RunOps = append(model.RunOps, func(run *domain.Run) { (*run).SetColor(*color) })
		}
	},

	css.BackgroundColor: func(value css.StyleValue, model *StyleModel) {
		var color *domain.Color
		if color = converters.ConvertColor(string(value)); color == nil {
			return
		}
		if highlightcolor := converters.ConvertToHighlightColor(*color); highlightcolor != domain.HighlightNone {
			model.RunOps = append(model.RunOps, func(run *domain.Run) { (*run).SetHighlight(highlightcolor) })
		}
		model.CellOps = append(model.CellOps, func(cell *domain.TableCell) { (*cell).SetShading(*color) })
	},

	css.Background: func(value css.StyleValue, model *StyleModel) {
		var color *domain.Color
		if color = converters.ConvertBackground(value); color == nil {
			return
		}
		if highlightcolor := converters.ConvertToHighlightColor(*color); highlightcolor != domain.HighlightNone {
			model.RunOps = append(model.RunOps, func(run *domain.Run) { (*run).SetHighlight(highlightcolor) })
		}
		model.CellOps = append(model.CellOps, func(cell *domain.TableCell) { (*cell).SetShading(*color) })
	},

	css.FontWeight: func(value css.StyleValue, model *StyleModel) {
		if *converters.ConvertBold(value) {
			model.RunOps = append(model.RunOps, func(run *domain.Run) { (*run).SetBold(true) })
		}
	},

	css.FontFamily: func(value css.StyleValue, model *StyleModel) {
		if font := string(value); font != "" {
			model.RunOps = append(model.RunOps, func(run *domain.Run) { (*run).SetFont(domain.Font{Name: font}) })
		}
	},

	css.FontStyle: func(value css.StyleValue, model *StyleModel) {
		if *converters.ConvertItalic(value) {
			model.RunOps = append(model.RunOps, func(run *domain.Run) { (*run).SetItalic(true) })
		}
	},

	css.FontSize: func(value css.StyleValue, model *StyleModel) {
		if size := converters.ConvertFontSize(value); size != nil {
			model.RunOps = append(model.RunOps, func(run *domain.Run) { (*run).SetSize(*size) })
		}
	},

	css.TextDecoration: func(value css.StyleValue, model *StyleModel) {
		switch value {
		case css.Underline:
			model.RunOps = append(model.RunOps, func(run *domain.Run) { (*run).SetUnderline(domain.UnderlineSingle) })
		case css.LineThrough:
			model.RunOps = append(model.RunOps, func(run *domain.Run) { (*run).SetStrike(true) })
		}
	},

	css.Width: func(value css.StyleValue, model *StyleModel) {
		if width := converters.ConvertWidth(value); width != nil {
			model.CellOps = append(model.CellOps, func(cell *domain.TableCell) { (*cell).SetWidth(*width) })
		}
	},

	css.BorderTop: func(value css.StyleValue, model *StyleModel) {
		model.Borders.Top = converters.ConvertBorder(string(value))
	},

	css.BorderLeft: func(value css.StyleValue, model *StyleModel) {
		model.Borders.Left = converters.ConvertBorder(string(value))
	},

	css.BorderRight: func(value css.StyleValue, model *StyleModel) {
		model.Borders.Right = converters.ConvertBorder(string(value))
	},

	css.BorderBottom: func(value css.StyleValue, model *StyleModel) {
		model.Borders.Bottom = converters.ConvertBorder(string(value))
	},

	css.Border: func(value css.StyleValue, model *StyleModel) {
		border := converters.ConvertBorder(string(value))
		model.Borders.Top = border
		model.Borders.Left = border
		model.Borders.Right = border
		model.Borders.Bottom = border
	},

	css.TextAlign: func(value css.StyleValue, model *StyleModel) {
		if align, ok := converters.TextAlignment[value]; ok {
			model.ParagraphOps = append(model.ParagraphOps, func(p *domain.Paragraph) { (*p).SetAlignment(align) })
		}
	},

	css.VerticalAlign: func(value css.StyleValue, model *StyleModel) {
		if align, ok := converters.VerticalAlignment[value]; ok {
			model.CellOps = append(model.CellOps, func(cell *domain.TableCell) { (*cell).SetVerticalAlignment(align) })
		}
		if align, ok := converters.VerticalAlignmentRun[value]; ok {
			model.RunOps = append(model.RunOps, func(run *domain.Run) { align(run, value) })
		}
	},

	css.TextTransform: func(value css.StyleValue, model *StyleModel) {
		model.TextTransformers = append(model.TextTransformers, TextTransforms[value])
	},

	css.TextDecorationStyle: func(value css.StyleValue, model *StyleModel) {
		if decorationStyle, ok := converters.TextDecorationStyle[value]; ok && decorationStyle != domain.UnderlineNone {
			model.RunOps = append(model.RunOps, func(run *domain.Run) { (*run).SetUnderline(decorationStyle) })
		}
	},
	css.ListStyleType: func(value css.StyleValue, model *StyleModel) {
		if numID, ok := converters.ListStyleToNumID[value]; ok {
			model.ParagraphOps = append(model.ParagraphOps, func(p *domain.Paragraph) {
				ref := domain.NumberingReference{
					ID:    numID,
					Level: 0,
				}
				(*p).SetNumbering(ref)
			})
		}
	},
}
