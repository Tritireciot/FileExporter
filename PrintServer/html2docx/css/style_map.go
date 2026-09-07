package css

import (
	"PrintServer/PrintServer/html2docx/utils"
	"fmt"
	"strings"
)

type StyleMap struct {
	Styles map[StyleProperty]StyleValue
	Order  []StyleProperty
}

func (styleMap *StyleMap) Update(otherMap StyleMap) {
	for _, property := range otherMap.Order {
		styleMap.Order = append(styleMap.Order, property)
		styleMap.Styles[property] = otherMap.Styles[property]
	}
}

func (styleMap StyleMap) Copy() StyleMap {
	new_styles := utils.CopyMap(styleMap.Styles)
	new_order := utils.CopySlice(styleMap.Order)
	return StyleMap{Styles: new_styles, Order: new_order}
}

func (styleMap *StyleMap) String() string {
	var styleStr strings.Builder
	for _, property := range styleMap.Order {
		fmt.Fprintf(&styleStr, "%s: %s;", property, styleMap.Styles[property])
	}
	return styleStr.String()
}
