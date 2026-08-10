package converters

import (
	"PrintServer/html2docx/css"
	"PrintServer/html2docx/utils"
	"strconv"

	"github.com/mmonterroca/docxgo/v2/domain"
)

var TextAlignment = map[css.StyleValue]domain.Alignment{
	css.LeftAlign:    domain.AlignmentLeft,
	css.RightAlign:   domain.AlignmentRight,
	css.EndAlign:     domain.AlignmentRight,
	css.JustifyAlign: domain.AlignmentJustify,
	css.CenterAlign:  domain.AlignmentCenter,
}

var VerticalAlignment = map[css.StyleValue]domain.VerticalAlignment{
	css.TopAlign:    domain.VerticalAlignTop,
	css.CenterAlign: domain.VerticalAlignCenter,
	css.BottomAlign: domain.VerticalAlignBottom,
}

type RunAlign func(run *domain.Run, value css.StyleValue)

var VerticalAlignmentRun = map[css.StyleValue]RunAlign{
	css.SubAlign: func(run *domain.Run, value css.StyleValue) {
		(*run).SetSubscript(value == css.SubAlign)
	},
	css.SuperAlign: func(run *domain.Run, value css.StyleValue) {
		(*run).SetSuperscript(value == css.SuperAlign)
	},
}

var TextDecorationStyle = map[css.StyleValue]domain.UnderlineStyle{
	css.Solid:     domain.UnderlineSingle,
	css.Dashed:    domain.UnderlineDashed,
	css.Double:    domain.UnderlineDouble,
	css.Dotted:    domain.UnderlineDotted,
	css.Wavy:      domain.UnderlineWave,
	css.NoneStyle: domain.UnderlineNone,
}

func ConvertBold(value css.StyleValue) *bool {
	if value == css.Bold {
		return utils.Pointer(true)
	}
	if weight, err := strconv.Atoi(string(value)); err == nil && weight >= 600 {
		return utils.Pointer(true)
	}
	return utils.Pointer(false)
}

func ConvertItalic(value css.StyleValue) *bool {
	if value != css.Normal {
		return utils.Pointer(true)
	}
	return utils.Pointer(false)
}

func ConvertFontSize(value css.StyleValue) *int {
	if size, err := strconv.Atoi(string(value)); err == nil {
		return &size
	}
	return nil
}
