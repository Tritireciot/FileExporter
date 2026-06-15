package consume

import (
	logging "PrintServer/agent"
	"PrintServer/mqutils"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
)

type messageReceivedCb func(m kafka.Message)

func logf(msg string, a ...interface{}) {
	fmt.Printf("Consuming error: "+msg, a...)
	fmt.Println()
}

func closeReader(reason string, topic string, r *kafka.Reader) {
	if err := r.Close(); err != nil {
		logging.Agent.AddSimpleError("Consuming", fmt.Sprintf("Closing during %s: failed to close reader: %s", reason, topic))
	} else {
		logging.Agent.AddSimpleInfo("Consuming", fmt.Sprintf("Closing during %s: %s closed!", reason, topic))
	}
}

// Чтение сообщение по определенной портиции
func Partition(ctx context.Context, topic string, partition int, brokers []string, wait *sync.WaitGroup, stop chan bool, cb messageReceivedCb) {
	wait.Add(1)
	defer wait.Done()

	// Запускаем цикл формирования объекта потребителя
	for {
		r := kafka.NewReader(kafka.ReaderConfig{
			Brokers:               brokers,
			Topic:                 topic,
			Partition:             partition,
			MinBytes:              10,
			MaxBytes:              10e6,                     // 10MB
			CommitInterval:        mqutils.CommitMQInterval, // flushes commits to Kafka every second
			WatchPartitionChanges: true,
			MaxWait:               mqutils.ConsumersReadTimeout,
			ErrorLogger:           kafka.LoggerFunc(logf),
		})

		logging.Agent.AddSimpleDebug("Starting", fmt.Sprintf("Consuming MQ(%v) at topic: %s(partiotion: %d)...", brokers, topic, partition))
		//----------------------------
		// Цикл обработки событий с рабочим потребителем
		for {
			m, err := r.ReadMessage(ctx)
			if err != nil {
				// Если происходит ошибка обработки сообщения - выходим из цикла потребления и производим открытие потребителя снова
				logging.Agent.AddSimpleError("Consuming", fmt.Sprintf("Location %s. Error: %v", topic, err))
				break
			} else {
				// Обработка сообщения
				cb(m)
			}

			// Завершаем работу по сигналу
			select {
			case <-stop:
				closeReader("exit", topic, r)
				return // Выходим из процесса чтения
			default:
			}
		}
		closeReader("error", topic, r)

		// Завершаем работу по сигналу
		select {
		case <-stop:
			closeReader("exit", topic, r)
			return // Выходим из процесса чтения
		default:
		}
	}
}

// Чтение сообщение в рамках группы
func Start(ctx context.Context, topic string, groupIDPref string, brokers []string, wait *sync.WaitGroup, stop chan bool, cb messageReceivedCb) {
	wait.Add(1)
	defer wait.Done()

	groupID := groupIDPref + topic // Замешиваем топик пока в клиенте проблема

	// Запускаем цикл формирования объекта потребителя
	for {
		r := kafka.NewReader(kafka.ReaderConfig{
			Brokers:               brokers,
			Topic:                 topic,
			GroupID:               groupID,
			MinBytes:              10,
			MaxBytes:              10e6,                     // 10MB
			CommitInterval:        mqutils.CommitMQInterval, // flushes commits to Kafka every second
			WatchPartitionChanges: true,
			MaxWait:               mqutils.ConsumersReadTimeout,
			ErrorLogger:           kafka.LoggerFunc(logf),
		})

		logging.Agent.AddSimpleDebug("Starting", fmt.Sprintf("Consuming MQ(%v) at topic: %s(groupID: %s)...", brokers, topic, groupID))
		//----------------------------
		// Цикл обработки событий с рабочим потребителем
		for {
			m, err := r.ReadMessage(ctx)
			if err != nil {
				// Если происходит ошибка обработки сообщения - выходим из цикла потребления и производим открытие потребителя снова
				logging.Agent.AddSimpleError("Consuming", fmt.Sprintf("Location %s. Error: %v", topic, err))
				break
			} else {
				// Обработка сообщения
				cb(m)
			}

			// Завершаем работу по сигналу
			select {
			case <-stop:
				closeReader("exit", topic, r)
				return // Выходим из процесса чтения
			default:
			}
		}
		// Закрываем потребителя в случае ошибки для котрытия его снова
		closeReader("error", topic, r)

		// Завершаем работу по сигналу
		select {
		case <-stop:
			return // Выходим из процесса чтения
		case <-time.After(5 * time.Second):
			// При ошибочном создании writer ожидаем 5 сек чтобы не было спама
		}
	}
}
