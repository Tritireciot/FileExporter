package logging

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const maxMessagesInQueue = 1024

// ------------------------------------------
// Доставка сообщений AutoplayX
type Delivery struct {
	queue   []Message
	mutex   sync.Mutex
	started bool // Флаг запуска доставщика

	SendFunc func(msg Message) error
}

// Добавление сообщения в очередь
func (d *Delivery) Add(m Message) {
	d.mutex.Lock()
	// Если у нас кол-во сообщений превышает допустимое значение - обрезаем самое старое
	// и добавляем новое в конец
	if len(d.queue) > maxMessagesInQueue {
		d.queue = d.queue[1:]
	}
	d.queue = append(d.queue, m)
	d.mutex.Unlock()
}

func (fd *Delivery) sendMsgsAndCut() {
	fd.mutex.Lock()
	// Обрабатываем пока есть события
	for {
		// Если присутствуют
		if len(fd.queue) > 0 {
			if err := fd.SendFunc(fd.queue[0]); err != nil {
				log.Print(err)
				break // ошибка ? выходим отпускаем очередь
			} else {
				// Отрезаем 1е отправленное сообщение
				fd.queue = fd.queue[1:]
			}
		} else {
			break // Если нет событий - выходим, отпускаем очередь
		}
	}
	fd.mutex.Unlock()
}

// Отправка сообщений типа AutoplayX
func (fd *Delivery) WriteFunc(wait *sync.WaitGroup, stop chan bool) {
	wait.Add(1)
	defer wait.Done()

	for {
		fd.sendMsgsAndCut()
		// Проверяем на необходимость выхода
		select {
		case <-time.After(2 * time.Second):
		case <-stop:
			// Обработаем все присутствующие и выйдем
			fd.sendMsgsAndCut()
			return
		}
	}
}

// Отправка сообщений типа AutoplayX в stdout
type StdoutDelivery struct {
	Delivery
}

func (sd *StdoutDelivery) Send(msg Message) error {
	text := fmt.Sprintf("%s ", severityTags[msg.Cat])

	keys := []MessageParamPreset{}
	for k := range msg.Params {
		// Не выводим данные Subsystem, Host, ProcessID, Application потому что мы и так о них знаем
		if k != Subsystem && k != Host && k != ProcessID && k != Application {
			keys = append(keys, k)
		}
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	for _, k := range keys {
		text += fmt.Sprintf(": %s(%v)", strings.ToUpper(messagePresetTags[k]), msg.Params[k])
	}

	log.Print(text)

	return nil
}

// Отправка сообщений типа AutoplayX в file
type FileDelivery struct {
	Delivery

	// Расположение файла лога
	path string
}

// Получить путь файла записи
func (fd *FileDelivery) fileName(subsystem string, app string) string {
	return filepath.Clean(fd.path + fmt.Sprintf("%s_%s_%s.log", subsystem, app, time.Now().Format("20060102")))
}

// Отправка лога в файл
func (fd *FileDelivery) Send(msg Message) error {
	var (
		subsystem, app string
	)
	subsystem, _ = msg.Params[Subsystem].(string)
	app, _ = msg.Params[Application].(string)

	// Открывый файл
	if f, err := os.OpenFile(fd.fileName(subsystem, app), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err != nil {
		log.Printf("can't open log file: %s : %v", fd.path, err)
		return err
	} else {
		if _, err := f.Write([]byte(msg.Text() + "\n\n")); err != nil {
			f.Close()
			log.Print(err)
		}
		if err := f.Close(); err != nil {
			log.Print(err)
		}
		return nil
	}
}
