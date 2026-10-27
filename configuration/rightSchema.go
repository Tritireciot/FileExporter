package config

const (
	// Функциональные права
	AccessAdministration  = 1001 // Редактивание серверов\служб
	AccessUserRights      = 1002 // Редактирование групп прав
	AccessUserRightsLinks = 1003 // Привязка групп прав к группам пользователей

)

// Объект и его права
type ObjectRightsSchema struct {
	Id   int           `json:"id,omitempty"`
	Type int           `json:"type,omitempty"`
	Name string        `json:"name,omitempty"`
	Cat  LangMapString `json:"category,omitempty"`

	ObjectsRights []SchemaParam `json:"rights,omitempty"`

	Objects []ObjectRightsSchema `json:"objects,omitempty"`
}
