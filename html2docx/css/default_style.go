package css

import "PrintServer/html2docx/html"

var DefaultTagStyles = map[html.Tag]StyleMap{
	html.StrongTag:  {FontWeight: Bold},
	html.BTag:       {FontWeight: Bold},
	html.EmTag:      {FontStyle: Italic},
	html.ITag:       {FontStyle: Italic},
	html.UTag:       {TextDecoration: Underline},
	html.InsTag:     {TextDecoration: Underline},
	html.DelTag:     {TextDecoration: LineThrough},
	html.STag:       {TextDecoration: LineThrough},
	html.SmallTag:   {FontSize: "20"},
	html.CodeTag:    {FontFamily: CourierNew, Color: CodeColor},
	html.LinkTag:    {Color: CssBlue, TextDecoration: Underline},
	html.Header1Tag: {FontSize: "44", FontWeight: Bold},
	html.Header2Tag: {FontSize: "36", FontWeight: Bold},
	html.Header3Tag: {FontSize: "32", FontWeight: Bold},
	html.Header4Tag: {FontSize: "28", FontWeight: Bold},
	html.Header5Tag: {FontSize: "24", FontWeight: Bold},
	html.Header6Tag: {FontSize: "24", FontWeight: Bold, FontStyle: Italic},
	html.TdTag:      {VerticalAlign: CenterAlign, TextAlign: CenterAlign},
	html.ThTag:      {VerticalAlign: CenterAlign, TextAlign: CenterAlign},
}
