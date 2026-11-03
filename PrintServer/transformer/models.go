package transformer

type Template struct {
	ID          int
	Name        string
	Content     string `json:"content"`
	Subsystem   string
	RenderData  map[string]bool
	Type        string `json:"template_type"`
	ContentHash string `json:"content_hash"`
	RevisionID  int    `json:"revision"`
}

type ValidateResult struct {
	Message string
	Payload any
}
