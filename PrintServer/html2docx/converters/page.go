package converters

import (
	"PrintServer/PrintServer/html2docx/css"

	"github.com/mmonterroca/docxgo/v2/domain"
)

type PageModel struct {
	Size        domain.PageSize
	Orientation domain.Orientation
	Margins     domain.Margins
}

func ConvertPaperSize(widthString, heightString string, widthSuffix, heightSuffix css.SizeUnit) *PageModel {
	var width, height *int
	if width := ConvertSize(widthString, widthSuffix); width == nil || *width == 0 {
		return nil
	}
	if height := ConvertSize(heightString, heightSuffix); height == nil || *height == 0 {
		return nil
	}

	page := PageModel{}
	if *width > *height {
		page.Orientation = domain.OrientationLandscape
	} else {
		page.Orientation = domain.OrientationPortrait
	}
	page.Size.Height = *height
	page.Size.Width = *width
	return &page

}

func ConvertPageSize(pageSize string) domain.PageSize {

	switch pageSize {
	case "A4":
		return domain.PageSizeA4
	case "A3":
		return domain.PageSizeA3
	case "letter":
		return domain.PageSizeLetter
	}
	return domain.PageSizeA4
}

func ConvertOrientation(orientation string) domain.Orientation {
	if orientation == "landscape" {
		return domain.OrientationLandscape
	}
	return domain.OrientationPortrait
}
