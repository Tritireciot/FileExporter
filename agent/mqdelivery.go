package logging

import (
	"PrintServer/mqutils"
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"
)

// Логирование сообщений в формате #19270
const LogEvents = "log.events"

// Отправка сообщений типа AutoplayX в mq
type MQDelivery struct {
	Delivery

	// Расположение очереди
	paths []string

	w *kafka.Writer
}

// Установим расположение очереди
func (d *MQDelivery) AXMQDelivery(ps []string) {
	d.paths = ps
}

func (fd *MQDelivery) Send(msg Message) error {
	// Создаем если не создан
	if fd.w == nil {
		fd.w = mqutils.CreateWriter(fd.paths, LogEvents, 0)
	}

	// Формируем тело сообщения
	ba, _ := json.Marshal(&msg)

	// Записываем сообщение в очередь
	if err := fd.w.WriteMessages(context.Background(), kafka.Message{Value: ba}); err != nil {
		return err
	}

	return nil
}
