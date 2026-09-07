package config

// --------------------------------------------------
type ServiceTypeID int

const (
	// Infrastructure
	ElasticSearchTypeID    ServiceTypeID = 1
	PostgreSQLSearchTypeID ServiceTypeID = 2
	AccessPointTypeID      ServiceTypeID = 3
	QueueTypeID            ServiceTypeID = 4
	AzimuthServiceTypeID   ServiceTypeID = 5
)

// Ключи нахождения общей конфигурации
const CFG_CERT_PUBLIC_PSQL = "cert.public"
const CFG_CERT_PRIVATE_PSQL = "cert.private"
const CFG_SUBSYSTEM_COMM_KEY_PSQL = "subsystem.communication.key"

// Сервисы
type ServiceProps struct {
	Name     string        `json:"name,omitempty"`     // Имя службы
	TypeID   ServiceTypeID `json:"typeId,omitempty"`   // Тип службы
	ServerID int           `json:"serverId,omitempty"` // ИД сервера на котором установлена службы
	Path     string        `json:"path,omitempty"`     // Путь по серверу (устанавливается только при запросе)
}

type Service struct {
	ID    int          `json:"id,omitempty"` // ИД службы
	Props ServiceProps `json:"props"`        // Свойства службы

	Configuration interface{} `json:"configuration,omitempty"` // Конфигурация службы
}

// --------------------------------------------------
// Сервер
type ServerProps struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type Server struct {
	Id    int         `json:"id"`
	Props ServerProps `json:"props"`
}

// Конфигурация Server-container
type ServerContainer struct {
	Server

	Services []Service `json:"services,omitempty"`
}

func (s *ServerContainer) Compare(sc *ServerContainer) bool {
	return s.Id == sc.Id && s.Props == sc.Props
}

// --------------------------------------------------------------------------------------
// Конфигурация сервиса при получении через GetServicesByType
type serviceConfiguration struct {
	Path          string      `json:"path"`
	Configuration interface{} `json:"configuration"`
}

type IdSlice struct {
	Ver int   `json:"ver,omitempty"`
	Ids []int `json:"ids"`
}
