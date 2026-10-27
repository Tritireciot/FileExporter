package pglistener

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/lib/pq"
)

// Listener структура для PostgreSQL listener
type Listener struct {
	config    Config
	listener  *pq.Listener
	handler   Handler
	stopChan  chan bool
	isRunning bool
	mu        sync.RWMutex
}

// New создание нового listener
func New(config Config, handler Handler) *Listener {
	return &Listener{
		config:   config,
		handler:  handler,
		stopChan: make(chan bool),
	}
}

// Start запуск listener
func (l *Listener) Start(channels ...string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.isRunning {
		return fmt.Errorf("listener is already running")
	}

	// Создаем строку подключения
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		l.config.Host, l.config.Port, l.config.User, l.config.Password, l.config.Database)

	// Создаем listener
	l.listener = pq.NewListener(connStr, 10*time.Second, time.Minute, func(ev pq.ListenerEventType, err error) {
		if err != nil {
			log.Printf("PostgreSQL listener event error: %v", err)
		}
	})

	// Подписываемся на каналы
	for _, channel := range channels {
		if err := l.listener.Listen(channel); err != nil {
			l.listener.Close()
			return fmt.Errorf("failed to listen on channel %s: %v", channel, err)
		}
		log.Printf("Subscribed to PostgreSQL channel: %s", channel)
	}

	l.isRunning = true

	// Запускаем в отдельной горутине
	go l.listenLoop(channels)

	return nil
}

// Stop остановка listener
func (l *Listener) Stop() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.isRunning {
		return nil
	}

	close(l.stopChan)

	if l.listener != nil {
		l.listener.Close()
	}

	// Закрываем обработчик
	if l.handler != nil {
		l.handler.Close()
	}

	l.isRunning = false
	return nil
}

// IsRunning проверка статуса
func (l *Listener) IsRunning() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.isRunning
}

// listenLoop основной цикл прослушивания
func (l *Listener) listenLoop(channels []string) {
	defer func() {
		l.mu.Lock()
		l.isRunning = false
		l.mu.Unlock()
	}()

	log.Printf("PostgreSQL listener started, monitoring channels: %v", channels)

	for {
		select {
		case notification := <-l.listener.Notify:
			if notification != nil {
				if l.handler != nil {
					if err := l.handler.HandleNotification(notification.Channel, notification.Extra); err != nil {
						log.Printf("Error handling notification: %v", err)
					}
				}
			}
		case <-l.stopChan:
			log.Println("PostgreSQL listener stopping...")
			return
		}
	}
}
