package parser

import (
	"PrintServer/PrintServer/html2docx/css"
	"fmt"
)

func Translate(currentRawStyles css.StyleMap, currentStyles *StyleModel) {
	for _, prop := range currentRawStyles.Order {
		if styleFunc, ok := parsers[prop]; ok {
			styleFunc(currentRawStyles.Styles[prop], currentStyles)
		} else {
			fmt.Println("None Parser for:", prop)
		}
	}

}
