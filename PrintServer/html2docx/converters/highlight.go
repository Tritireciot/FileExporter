package converters

import (
	"PrintServer/PrintServer/html2docx/css"

	"github.com/mmonterroca/docxgo/v2/domain"
)

func ConvertToHighlightColor(backgroundColor domain.Color) domain.HighlightColor {
	switch backgroundColor {
	case css.Yellow:
		return domain.HighlightYellow
	case css.DarkYellow:
		return domain.HighlightDarkYellow
	case css.Green:
		return domain.HighlightGreen
	case css.DarkGreen:
		return domain.HighlightDarkGreen
	case css.Cyan:
		return domain.HighlightCyan
	case css.DarkCyan:
		return domain.HighlightDarkCyan
	case css.Magenta:
		return domain.HighlightMagenta
	case css.DarkMagenta:
		return domain.HighlightDarkMagenta
	case css.Blue:
		return domain.HighlightBlue
	case css.DarkBlue:
		return domain.HighlightDarkBlue
	case css.Red:
		return domain.HighlightRed
	case css.DarkRed:
		return domain.HighlightDarkRed
	case css.LightGray:
		return domain.HighlightLightGray
	case css.DarkGray:
		return domain.HighlightDarkGray
	}
	return domain.HighlightNone
}
