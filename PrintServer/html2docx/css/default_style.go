package css

import "PrintServer/PrintServer/html2docx/html"

var DefaultTagStyles = map[html.Tag]StyleMap{
	html.StrongTag:  {Styles: map[StyleProperty]StyleValue{FontWeight: Bold}, Order: []StyleProperty{FontWeight}},
	html.BTag:       {Styles: map[StyleProperty]StyleValue{FontWeight: Bold}, Order: []StyleProperty{FontWeight}},
	html.EmTag:      {Styles: map[StyleProperty]StyleValue{FontStyle: Italic}, Order: []StyleProperty{FontStyle}},
	html.ITag:       {Styles: map[StyleProperty]StyleValue{FontStyle: Italic}, Order: []StyleProperty{FontStyle}},
	html.UTag:       {Styles: map[StyleProperty]StyleValue{TextDecoration: Underline}, Order: []StyleProperty{TextDecoration}},
	html.InsTag:     {Styles: map[StyleProperty]StyleValue{TextDecoration: Underline}, Order: []StyleProperty{TextDecoration}},
	html.DelTag:     {Styles: map[StyleProperty]StyleValue{TextDecoration: LineThrough}, Order: []StyleProperty{TextDecoration}},
	html.STag:       {Styles: map[StyleProperty]StyleValue{TextDecoration: LineThrough}, Order: []StyleProperty{TextDecoration}},
	html.SmallTag:   {Styles: map[StyleProperty]StyleValue{FontSize: "20"}, Order: []StyleProperty{FontSize}},
	html.CodeTag:    {Styles: map[StyleProperty]StyleValue{FontFamily: CourierNew, Color: CodeColor}, Order: []StyleProperty{FontFamily, Color}},
	html.LinkTag:    {Styles: map[StyleProperty]StyleValue{TextDecoration: Underline, Color: CssBlue}, Order: []StyleProperty{TextDecoration, Color}},
	html.Header1Tag: {Styles: map[StyleProperty]StyleValue{FontSize: "44", FontWeight: Bold}, Order: []StyleProperty{FontSize, FontWeight}},
	html.Header2Tag: {Styles: map[StyleProperty]StyleValue{FontSize: "36", FontWeight: Bold}, Order: []StyleProperty{FontSize, FontWeight}},
	html.Header3Tag: {Styles: map[StyleProperty]StyleValue{FontSize: "32", FontWeight: Bold}, Order: []StyleProperty{FontSize, FontWeight}},
	html.Header4Tag: {Styles: map[StyleProperty]StyleValue{FontSize: "28", FontWeight: Bold}, Order: []StyleProperty{FontSize, FontWeight}},
	html.Header5Tag: {Styles: map[StyleProperty]StyleValue{FontSize: "24", FontWeight: Bold}, Order: []StyleProperty{FontSize, FontWeight}},
	html.Header6Tag: {Styles: map[StyleProperty]StyleValue{FontSize: "24", FontWeight: Bold, FontStyle: Italic}, Order: []StyleProperty{FontSize, FontWeight, FontStyle}},
	html.TdTag:      {Styles: map[StyleProperty]StyleValue{VerticalAlign: CenterAlign, TextAlign: LeftAlign}, Order: []StyleProperty{VerticalAlign, TextAlign}},
	html.ThTag:      {Styles: map[StyleProperty]StyleValue{VerticalAlign: CenterAlign, TextAlign: CenterAlign}, Order: []StyleProperty{VerticalAlign, TextAlign}},
	html.SupTag:     {Styles: map[StyleProperty]StyleValue{VerticalAlign: SuperAlign}, Order: []StyleProperty{VerticalAlign}},
	html.SubTag:     {Styles: map[StyleProperty]StyleValue{VerticalAlign: SubAlign}, Order: []StyleProperty{VerticalAlign}},
}
