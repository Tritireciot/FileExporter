package parser

import (
	"PrintServer/html2docx/css"
	"fmt"
)

func Translate(currentRawStyles css.StyleMap, currentStyles *StyleModel) {
	for prop, value := range currentRawStyles {
		if styleFunc, ok := parsers[prop]; ok {
			styleFunc(value, currentStyles)
		} else {
			fmt.Println("None Parser for:", prop)
		}
	}

}
