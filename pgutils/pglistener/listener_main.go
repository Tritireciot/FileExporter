package pglistener

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	//"sync"
)

// RunAsync запускает listener асинхронно с корректной остановкой
func RunAsync(config Config, handler Handler, stop chan bool, channels ...string) *Listener {
	// Создаем listener
	listener := New(config, handler)

	// Запускаем listener
	if err := listener.Start(channels...); err != nil {
		log.Printf("Failed to start PostgreSQL listener: %v", err)
		return listener
	}

	log.Println("PostgreSQL listener started successfully")

	// Горутина для остановки listener при получении сигнала stop
	go func() {
		<-stop
		log.Println("Stopping PostgreSQL listener...")
		if listener != nil {
			listener.Stop()
		}
	}()

	return listener
}

// RunWithWaitAndStop запускает listener и ожидает сигнал остановки
func RunWithWaitAndStop(config Config, handler Handler, stop chan bool, channels ...string) *Listener {
	// Создаем listener
	listener := New(config, handler)

	// Запускаем listener
	if err := listener.Start(channels...); err != nil {
		log.Printf("Failed to start PostgreSQL listener: %v", err)
		return listener
	}

	log.Println("PostgreSQL listener started successfully")

	// Ждем сигнал остановки в отдельной горутине
	go func() {
		<-stop
		log.Println("Stopping PostgreSQL listener...")
		if listener != nil {
			listener.Stop()
		}
	}()

	return listener
}

// Run запускает PostgreSQL listener с заданным обработчиком
func Run(config Config, handler Handler, channels ...string) error {
	// Создаем listener
	listener := New(config, handler)

	// Запускаем listener
	if err := listener.Start(channels...); err != nil {
		return err
	}

	log.Println("PostgreSQL listener is running...")

	// Ожидание сигнала завершения
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Ждем сигнал завершения
	sig := <-sigChan
	log.Printf("Received signal: %v", sig)

	// Останавливаем listener
	if err := listener.Stop(); err != nil {
		log.Printf("Error stopping listener: %v", err)
	}

	log.Println("PostgreSQL listener stopped")
	return nil
}

// RunWithWait запускает listener и ожидает сигнал завершения
func RunWithWait(config Config, handler Handler, channels ...string) {
	// Создаем listener
	listener := New(config, handler)

	// Запускаем listener
	if err := listener.Start(channels...); err != nil {
		log.Fatalf("Failed to start PostgreSQL listener: %v", err)
	}

	log.Println("PostgreSQL listener is running...")

	// Ожидание сигнала завершения
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Ждем сигнал завершения
	sig := <-sigChan
	log.Printf("Received signal: %v", sig)

	// Останавливаем listener
	if err := listener.Stop(); err != nil {
		log.Printf("Error stopping listener: %v", err)
	}

	log.Println("PostgreSQL listener stopped")
}
