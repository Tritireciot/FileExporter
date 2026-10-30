package template_compiler

import "PrintServer/PrintServer/diagnostic"

type ParseResult struct {
	Output      string
	Diagnostics []diagnostic.Diagnostic
}

func (r ParseResult) HasErrors() bool {
	return len(r.Diagnostics) != 0
}
