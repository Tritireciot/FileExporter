package css

import "github.com/mmonterroca/docxgo/v2/domain"

type GlobalStyles struct {
	Classes      map[string]StyleMap
	IDs          map[string]StyleMap
	Tags         map[string]StyleMap
	Page         PageModel
	DefaultStyle StyleMap
}

type PageModel struct {
	Size        domain.PageSize
	Orientation domain.Orientation
	Margins     domain.Margins
}
