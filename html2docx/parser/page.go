package parser

import (
	"PrintServer/html2docx/converters"
	"PrintServer/html2docx/css"
	"strings"
)

func parsePageRule(templateStyles *css.GlobalStyles, props map[string]string) {

	for property, value := range props {
		switch css.StyleProperty(property) {
		case css.Size:
			parsePageSize(templateStyles, value)
		case css.Margin:
			parsePageMargin(templateStyles, value)
		case css.MarginTop:
			if margin := converters.ConvertSize(value, converters.CheckSizeSuffix(value)); margin != nil {
				templateStyles.Page.Margins.Top = *margin
			}
		case css.MarginBottom:
			if margin := converters.ConvertSize(value, converters.CheckSizeSuffix(value)); margin != nil {
				templateStyles.Page.Margins.Bottom = *margin
			}
		case css.MarginRight:
			if margin := converters.ConvertSize(value, converters.CheckSizeSuffix(value)); margin != nil {
				templateStyles.Page.Margins.Right = *margin
			}
		case css.MarginLeft:
			if margin := converters.ConvertSize(value, converters.CheckSizeSuffix(value)); margin != nil {
				templateStyles.Page.Margins.Left = *margin
			}
		}

	}

}

func parsePageSize(templateStyles *css.GlobalStyles, value string) {
	size_parts := strings.Split(value, " ")
	if len(size_parts) == 2 {
		suffix1 := converters.CheckSizeSuffix(size_parts[0])
		suffix2 := converters.CheckSizeSuffix(size_parts[1])
		if suffix1 != css.NONE_UNIT && suffix2 != css.NONE_UNIT {
			temp_page := converters.ConvertPaperSize(size_parts[0], size_parts[1], suffix1, suffix2)
			if temp_page.Size.Height > 0 && temp_page.Size.Width > 0 {
				templateStyles.Page.Size = temp_page.Size
				templateStyles.Page.Orientation = temp_page.Orientation
			}
		} else {
			templateStyles.Page.Size = converters.ConvertPageSize(size_parts[0])
			templateStyles.Page.Orientation = converters.ConvertOrientation(size_parts[1])
		}

	} else if len(size_parts) == 1 {
		page_size := converters.ConvertPageSize(size_parts[0])
		if page_size.Width == page_size.Height && page_size.Height == 0 {
			templateStyles.Page.Orientation = converters.ConvertOrientation(size_parts[0])
		} else {
			templateStyles.Page.Size = page_size
		}
	}

}

func parsePageMargin(templateStyles *css.GlobalStyles, value string) {
	margin_parts := strings.Split(value, " ")
	if len(margin_parts) == 4 {
		suffixs := []string{}
		for _, margin := range margin_parts {
			suffixs = append(suffixs, string(converters.CheckSizeSuffix(margin)))
		}
		templateStyles.Page.Margins = converters.ConvertMargins(margin_parts, suffixs)
	} else if len(margin_parts) == 1 {
		templateStyles.Page.Margins = converters.ConvertMargins([]string{value}, []string{string(converters.CheckSizeSuffix(value))})
	}
}
