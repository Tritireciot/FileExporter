package transformer

type Template struct {
	ID int
	Name string
	Content string
	Subsystem string
	RenderData map[string]bool
}