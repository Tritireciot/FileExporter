package css

import "github.com/mmonterroca/docxgo/v2/domain"

type GlobalStyles struct {
	Classes       ClassRule
	IDs           map[string]StyleMap
	Tags          map[string]StyleMap
	Page          PageModel
	DefaultStyle  StyleMap
	ComplexStyles []*StyleNode
}

type ClassRule struct {
	Order       []string
	ClassStyles map[string]StyleMap
}
type StyleNode struct {
	Parent       *StyleNode
	Selector     string
	StyleContent StyleMap
	Modifier     string
	Connection   string
}

func (sn *StyleNode) String() string {
	if sn == nil {
		return ""
	}
	return sn.Selector + " Connect: \"" + sn.Connection + "\" Parent: " + sn.Parent.String()
}

type PageModel struct {
	Size        domain.PageSize
	Orientation domain.Orientation
	Margins     domain.Margins
}
