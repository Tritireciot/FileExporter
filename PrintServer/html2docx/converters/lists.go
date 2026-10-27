package converters

import (
	"PrintServer/PrintServer/html2docx/css"
)

var ListStyleToNumID = map[css.StyleValue]int{
	css.ListStyleDisc:               1, // abstractNumId="1" -> Точки
	css.ListStyleSquare:             4, // abstractNumId="2" -> Квадраты
	css.ListStyleDecimal:            5, // abstractNumId="3" -> Обычные цифры
	css.ListStyleCircle:             3, // abstractNumId="4" -> Пустые кружки
	css.ListStyleDecimalLeadingZero: 6, // abstractNumId="5" -> 01, 02
	css.ListStyleUpperRoman:         7, // abstractNumId="6" -> I, II, III
	css.ListStyleLowerAlpha:         8, // abstractNumId="7" -> a, b, c
	css.ListStyleUpperAlpha:         9, // abstractNumId="8" -> A, B, C
}
