package transformer

type Template struct {
	ID         int
	Name       string
	Content    string
	Subsystem  string
	RenderData map[string]bool
	Type       string
}

type ValidateResult struct {
	Message string
	Payload any
}

