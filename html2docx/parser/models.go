package parser

import (
	"PrintServer/html2docx/css"
	"PrintServer/html2docx/utils"
	"strings"

	"github.com/mmonterroca/docxgo/v2/domain"
)

type TagStyleState struct {
	ElementModel StyleModel
	Link         string
	TextPrefix   string
	Order        int
}

type CellData struct {
	Data      string
	Row       int
	Col       int
	Merge     []int
	Skip      bool
	CellStyle StyleModel
	TextStyle css.StyleMap
}

type RunModifier func(run *domain.Run)
type ParagraphModifier func(p *domain.Paragraph)
type TableModifier func(t *domain.Table)
type CellModifier func(c *domain.TableCell)
type TextTransform func(str string) string

type StyleModel struct {
	RunOps       []RunModifier
	ParagraphOps []ParagraphModifier
	TableOps     []TableModifier
	CellOps      []CellModifier

	Borders domain.TableBorders

	TextPrefix       string
	TextTransformers []TextTransform
}

func (m StyleModel) Clone() StyleModel {
	cloned := m

	if m.RunOps != nil {
		cloned.RunOps = make([]RunModifier, len(m.RunOps))
		copy(cloned.RunOps, m.RunOps)
	}

	if m.ParagraphOps != nil {
		cloned.ParagraphOps = make([]ParagraphModifier, len(m.ParagraphOps))
		copy(cloned.ParagraphOps, m.ParagraphOps)
	}

	if m.CellOps != nil {
		cloned.CellOps = make([]CellModifier, len(m.CellOps))
		copy(cloned.CellOps, m.CellOps)
	}

	if m.TextTransformers != nil {
		cloned.TextTransformers = make([]TextTransform, len(m.TextTransformers))
		copy(cloned.TextTransformers, m.TextTransformers)
	}

	return cloned
}

func (m *StyleModel) ApplyToRun(r *domain.Run) {
	for _, apply := range m.RunOps {
		apply(r)
	}
}

func (m *StyleModel) ApplyToTable(t *domain.Table) {
	for _, apply := range m.TableOps {
		apply(t)
	}
}

func (m *StyleModel) ApplyToParagraph(p *domain.Paragraph) {
	for _, apply := range m.ParagraphOps {
		apply(p)
	}
}

func (m *StyleModel) ApplyToCell(cell *domain.TableCell) {
	for _, apply := range m.CellOps {
		apply(cell)
	}
	(*cell).SetBorders(m.Borders)
}

func (m *StyleModel) TransformText(text string) string {
	for _, transform := range m.TextTransformers {
		text = transform(text)
	}
	return m.TextPrefix + text
}

var TextTransforms = map[css.StyleValue]TextTransform{
	css.UpperCase: func(str string) string {
		return strings.ToUpper(str)
	},

	css.LowerCase: func(str string) string {
		return strings.ToLower(str)
	},

	css.Capitalize: func(str string) string {
		return utils.Capitalize(str)
	},

	css.NoneStyle: func(str string) string {
		return str
	},
}

func BuildStyleModel(rawStyles css.StyleMap) StyleModel {
	var model StyleModel

	for property, value := range rawStyles {

		if parseFunc, exists := parsers[property]; exists {
			parseFunc(value, &model)
		}
	}

	return model
}
