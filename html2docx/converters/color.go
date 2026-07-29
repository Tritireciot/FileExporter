package converters

import (
	"PrintServer/html2docx/css"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/mmonterroca/docxgo/v2/domain"
)

func ConvertColor(value string) *domain.Color {
	if strings.HasPrefix(value, "#") {
		color_code := strings.TrimPrefix(value, "#")
		colorLen := utf8.RuneCountInString(color_code)
		var R, G, B uint8
		if 3 <= colorLen && colorLen <= 4 {
			red, _ := strconv.ParseInt(string(color_code[0])+string(color_code[0]), 16, 64)
			R = uint8(red)
			green, _ := strconv.ParseInt(string(color_code[1])+string(color_code[1]), 16, 64)
			G = uint8(green)
			blue, _ := strconv.ParseInt(string(color_code[2])+string(color_code[2]), 16, 64)
			B = uint8(blue)
		} else if 6 <= colorLen && colorLen <= 8 {
			red, _ := strconv.ParseInt(string(color_code[0:2]), 16, 64)
			R = uint8(red)
			green, _ := strconv.ParseInt(string(color_code[2:4]), 16, 64)
			G = uint8(green)
			blue, _ := strconv.ParseInt(string(color_code[4:6]), 16, 64)
			B = uint8(blue)
		}
		return &domain.Color{R: R, G: G, B: B}
	} else if strings.HasPrefix(value, "rgb") || strings.HasPrefix(value, "rgba") {
		if strings.HasPrefix(value, "rgba") {
			value = strings.TrimPrefix(value, "rgba(")
		} else {
			value = strings.TrimPrefix(value, "rgb(")
		}
		value = strings.TrimSuffix(value, ")")
		var colors []string
		if strings.Contains(value, ",") {
			colors = strings.Split(value, ",")
		} else {
			colors = strings.Split(value, " ")
		}
		red, _ := strconv.ParseInt(colors[0], 10, 64)
		R := uint8(red)
		green, _ := strconv.ParseInt(colors[1], 10, 64)
		G := uint8(green)
		blue, _ := strconv.ParseInt(colors[2], 10, 64)
		B := uint8(blue)
		return &domain.Color{R: R, G: G, B: B}
	} else {
		switch value {
		case "black":
			return &domain.ColorBlack
		case "white":
			return &domain.ColorWhite
		case "green":
			return &domain.ColorGreen
		case "blue":
			return &domain.ColorBlue
		case "red":
			return &domain.ColorRed
		case "yellow":
			return &css.Yellow
		}

	}
	return nil
}
