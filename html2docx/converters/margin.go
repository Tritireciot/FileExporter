package converters

import (
	"PrintServer/html2docx/css"

	"github.com/mmonterroca/docxgo/v2/domain"
)

func ConvertMargins(margins, suffixs []string) domain.Margins {
	page_margin := domain.DefaultMargins
	if len(margins) == 1 {
		margin := ConvertSize(margins[0], css.SizeUnit(suffixs[0]))
		if margin != nil {
			page_margin.Left = *margin
			page_margin.Right = *margin
			page_margin.Top = *margin
			page_margin.Bottom = *margin
		}
	} else {
		if left := ConvertSize(margins[0], css.SizeUnit(suffixs[0])); left != nil {
			page_margin.Left = *left
		}
		if top := ConvertSize(margins[1], css.SizeUnit(suffixs[1])); top != nil {
			page_margin.Top = *top
		}
		if right := ConvertSize(margins[2], css.SizeUnit(suffixs[2])); right != nil {
			page_margin.Right = *right
		}
		if bottom := ConvertSize(margins[3], css.SizeUnit(suffixs[3])); bottom != nil {
			page_margin.Bottom = *bottom
		}

	}
	return page_margin
}
