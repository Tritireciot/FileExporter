package css

import "github.com/mmonterroca/docxgo/v2/domain"

// Единицы измерения

type SizeUnit string

const (
	NONE_UNIT SizeUnit = ""
	MM        SizeUnit = "mm"
	CM        SizeUnit = "cm"
	IN        SizeUnit = "in"
	PT        SizeUnit = "pt"
	PC        SizeUnit = "pc"
	PX        SizeUnit = "px"
)

var PageSizeUnits []SizeUnit = []SizeUnit{
	MM, CM, IN, PT, PC, PX,
}

// Highlight Colors

var (
	Yellow      domain.Color = domain.Color{R: 255, G: 255, B: 0}
	DarkYellow  domain.Color = domain.Color{R: 139, G: 139, B: 0}
	Green       domain.Color = domain.Color{R: 0, G: 255, B: 0}
	DarkGreen   domain.Color = domain.Color{R: 0, G: 139, B: 0}
	Cyan        domain.Color = domain.Color{R: 0, G: 255, B: 255}
	DarkCyan    domain.Color = domain.Color{R: 0, G: 139, B: 139}
	Magenta     domain.Color = domain.Color{R: 255, G: 0, B: 255}
	DarkMagenta domain.Color = domain.Color{R: 139, G: 0, B: 139}
	Blue        domain.Color = domain.Color{R: 0, G: 0, B: 255}
	DarkBlue    domain.Color = domain.Color{R: 0, G: 0, B: 139}
	Red         domain.Color = domain.Color{R: 255, G: 0, B: 0}
	DarkRed     domain.Color = domain.Color{R: 139, G: 0, B: 0}
	LightGray   domain.Color = domain.Color{R: 211, G: 211, B: 211}
	DarkGray    domain.Color = domain.Color{R: 169, G: 169, B: 169}
)

// Styles Properties

type StyleProperty string

const (
	FontWeight          StyleProperty = "font-weight"
	FontStyle           StyleProperty = "font-style"
	TextDecoration      StyleProperty = "text-decoration"
	FontSize            StyleProperty = "font-size"
	FontFamily          StyleProperty = "font-family"
	Color               StyleProperty = "color"
	BackgroundColor     StyleProperty = "background-color"
	Background          StyleProperty = "background"
	Width               StyleProperty = "width"
	BorderTop           StyleProperty = "border-top"
	BorderLeft          StyleProperty = "border-left"
	BorderRight         StyleProperty = "border-right"
	BorderBottom        StyleProperty = "border-bottom"
	Border              StyleProperty = "border"
	TextAlign           StyleProperty = "text-align"
	VerticalAlign       StyleProperty = "vertical-align"
	TextTransform       StyleProperty = "text-transform"
	TextDecorationStyle StyleProperty = "text-decoration-style"
	LineHeight          StyleProperty = "line-height"
	Page                StyleProperty = "@page"
	Margin              StyleProperty = "margin"
	MarginTop           StyleProperty = "margin-top"
	MarginBottom        StyleProperty = "margin-bottom"
	MarginRight         StyleProperty = "margin-right"
	MarginLeft          StyleProperty = "margin-left"
	Size                StyleProperty = "size"
	ListStyleType       StyleProperty = "list-style-type"
)

// Styles Values

type StyleValue string

// Default Values

const (
	NoneStyle StyleValue = "none"
	Normal    StyleValue = "normal"
)

// Weight
const (
	Bold StyleValue = "bold"
)

// Font Style
const (
	Italic StyleValue = "italic"
)

// Text Decoration
const (
	Underline   StyleValue = "underline"
	LineThrough StyleValue = "line-through"
)

// Text Decoration Style
const (
	Solid  StyleValue = "solid"
	Dashed StyleValue = "dashed"
	Wavy   StyleValue = "wavy"
	Dotted StyleValue = "dotted"
	Double StyleValue = "double"
)

// Font Families
const (
	Arial      StyleValue = "Arial"
	CourierNew StyleValue = "Courier New"
)

// Colors
const (
	CssBlue   StyleValue = "blue"
	CodeColor StyleValue = "rgb(163,21,21)"
)

// Text Alignment
const (
	LeftAlign    StyleValue = "left"
	RightAlign   StyleValue = "right"
	EndAlign     StyleValue = "end"
	JustifyAlign StyleValue = "justify"
	CenterAlign  StyleValue = "center"
	TopAlign     StyleValue = "top"
	BottomAlign  StyleValue = "bottom"

	SuperAlign StyleValue = "super"
	SubAlign   StyleValue = "sub"
)

// Text Transform

const (
	UpperCase  StyleValue = "uppercase"
	LowerCase  StyleValue = "lowercase"
	Capitalize StyleValue = "capitalize"
)

// List Style Type

const (
	ListStyleDisc               StyleValue = "disc"
	ListStyleCircle             StyleValue = "circle"
	ListStyleSquare             StyleValue = "square"
	ListStyleDecimal            StyleValue = "decimal"
	ListStyleDecimalLeadingZero StyleValue = "decimal-leading-zero"
	ListStyleUpperRoman         StyleValue = "upper-roman"
	ListStyleLowerAlpha         StyleValue = "lower-alpha"
	ListStyleUpperAlpha         StyleValue = "upper-alpha"
)
