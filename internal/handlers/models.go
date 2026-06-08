package handlers

type ExportModel struct {
	TemplateId int            `json:"template_id"`
	Data       any 			  `json:"data"`
	Subsystem string 		  `json:"subsystem"`
}


func (model *ExportModel) Validate() bool {
	return model.TemplateId > 0 && len(model.Subsystem) > 0
}


type IDModel struct {
	ID int `json:"ID"`
}