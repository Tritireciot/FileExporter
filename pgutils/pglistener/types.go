package pglistener

// Config конфигурация для подключения к PostgreSQL
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

// Notification структура для уведомлений из PostgreSQL
type Notification struct {
	Topic     string      `json:"topic"`
	Message   interface{} `json:"message"`
	Timestamp string      `json:"timestamp"`
}
