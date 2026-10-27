package converters

import (
	"PrintServer/PrintServer/html2docx/css"
	"PrintServer/PrintServer/html2docx/utils"
	"math"
	"strconv"
	"strings"
)

func ConvertSize(distStr string, suffix css.SizeUnit) *int {
	dist, err := strconv.ParseFloat(strings.TrimSuffix(distStr, string(suffix)), 64)
	if err != nil {
		return nil
	}
	switch suffix {
	case css.MM:
		return utils.Pointer(int(math.Round(dist * (1440.0 / 25.4))))
	case css.CM:
		return utils.Pointer(int(math.Round(dist * (1440.0 / 2.54))))
	case css.IN:
		return utils.Pointer(int(math.Round(dist * 1440.0)))
	case css.PT:
		return utils.Pointer(int(math.Round(dist * 20.0)))
	case css.PC:
		return utils.Pointer(int(math.Round(dist * 240.0)))
	case css.PX:
		return utils.Pointer(int(math.Round(dist * 15.0)))
	case css.NONE_UNIT:
		return utils.Pointer(int(math.Round(dist * 2)))
	}
	return nil
}

func CheckSizeSuffix(sizeString string) css.SizeUnit {
	for _, size_suffix := range css.PageSizeUnits {
		if strings.HasSuffix(sizeString, string(size_suffix)) {
			return size_suffix
		}
	}
	return css.NONE_UNIT
}

func ConvertWidth(value css.StyleValue) *int {
	size_suffix := CheckSizeSuffix(string(value))
	return ConvertSize(
		strings.TrimSuffix(string(value), string(size_suffix)),
		size_suffix,
	)
}
