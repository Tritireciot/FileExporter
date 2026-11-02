package diagnostic

type Location struct {
	Line   int `json:"line,omitempty"`
	Column int `json:"column,omitempty"`
}

type Span struct {
	Start Location `json:"start"`
	End   Location `json:"end"`
}
