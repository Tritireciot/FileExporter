package converters

import (
	"PrintServer/html2docx/css"
	"strconv"
	"strings"

	"github.com/mmonterroca/docxgo/v2/domain"
)

func getLineStyle(value string) domain.BorderLineStyle {
	switch value {
	case "dashed":
		return domain.BorderDashed
	case "none":
		return domain.BorderNone
	}
	return domain.BorderSingle
}

func ConvertBorder(borderStr string) domain.BorderStyle {
	style := domain.BorderSingle
	width := 0
	color := domain.ColorWhite
	components := strings.SplitN(borderStr, " ", 3)
	switch len(components) {
	case 1:
		style = getLineStyle(components[0])
	case 2:
		setted := false
		for _, sizeSuffix := range css.PageSizeUnits {
			if strings.HasSuffix(components[0], string(sizeSuffix)) {
				size := strings.TrimSuffix(components[0], string(sizeSuffix))
				width, _ = strconv.Atoi(size)
				width *= 8
				style = getLineStyle(components[1])
				setted = true
				break
			}
		}
		if !setted {
			style = getLineStyle(components[0])
			if convertedColor := ConvertColor(components[1]); convertedColor != nil {
				color = *convertedColor
			}
		}
	case 3:
		for _, sizeSuffix := range css.PageSizeUnits {
			if strings.HasSuffix(components[0], string(sizeSuffix)) {
				size := strings.TrimSuffix(components[0], string(sizeSuffix))
				width, _ = strconv.Atoi(size)
				width *= 8
				break
			}
		}
		style = getLineStyle(components[1])
		if convertedColor := ConvertColor(components[2]); convertedColor != nil {
			color = *convertedColor
		}

	}
	return domain.BorderStyle{
		Style: style,
		Width: width,
		Color: color,
	}
}
