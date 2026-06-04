package handlers

type ExportModel struct {
	TemplateId int            `json:"template_id"`
	Data       map[string]any `json:"data"`
	Subsystem string 		  `json:"susbsystem"`
}

func (model *ExportModel) Validate() bool {
	return len(model.Data) > 0
}


type IDModel struct {
	ID int `json:"ID"`
}