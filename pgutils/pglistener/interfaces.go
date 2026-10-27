package pglistener

// Handler интерфейс для обработки уведомлений
type Handler interface {
	HandleNotification(channel string, payload string) error
	Close() error
}
