package logging

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Возможные назначения логов
type Destinations struct {
	Stdout bool // Производить вывод в stdout
	File   bool // Производить вывод в файл
	Mq     bool // Производить вывод в очередь
}

func (d *Destinations) Any() {
	d.Stdout = true
	d.File = true
	d.Mq = true
}

type Config struct {
	Destinations

	FileLogPath string // Путь до логов файла
}

// Объект через который производится логирование
type axAgent struct {
	file   FileDelivery
	stdout StdoutDelivery
	mq     MQDelivery

	Config Config // Параметры работы

	// Текущее расположение из которого производится логирование
	Location AXLocation
}

var (
	Agent *axAgent // Глобальная переменная агента для удобства
)

// Инициализировать глобальную перменную
// config - первоначальная конфигурация агента
func InitAXAgent(config Config, location AXLocation) {
	if Agent == nil {
		Agent = &axAgent{Config: config, Location: location}
		envpath := os.Getenv("LOG_PATH")
		if len(envpath) > 0 {
			Agent.file.path = envpath + string(filepath.Separator)
		}
	}
}

// Создать агент вручную
func NewAXAgent(config Config, location AXLocation) *axAgent {
	agent := &axAgent{Config: config, Location: location}
	envpath := os.Getenv("LOG_PATH")
	if len(envpath) > 0 {
		Agent.file.path = envpath + string(filepath.Separator)
	}
	return agent
}

//-----------------------------------------------------------------------
// Перед использованием:
// * Инициализируйте/Создайте объект логгера с нужной конфигурацией
//		logging.InitAXAgent(logging.Config{Stdout: true, File: true, Mq: true}, logging.AXLocation{Subsystem: "Air", App: "App", Host: host, ProcessId: uint32(os.Getpid())})
// * Если нужно отправлять в очередь сообщений - установите SetMQPaths
// * Запустите потоки логирования командой Start
// Например:
//	logging.Agent.Start(wait, stop)
// wait и stop тут объекты синхронизации для корректного завершения
//-----------------------------------------------------------------------

// Установить пути до очереди сообщений - отключение опции очереди если она не указана
func (a *axAgent) SetMQPaths(ps []string) {
	a.mq.paths = ps

	if len(a.mq.paths) <= 0 {
		a.Config.Mq = false
		log.Print("Disable logger MQ option becouse of empty MQ configuration!")
	} else {
		a.Config.Mq = true
		log.Printf("Enable mq logging feature with: %v.", a.mq.paths)
	}
}

// Запустить обрабатывающие потоки
func (a *axAgent) Start(wait *sync.WaitGroup, stop chan bool) {
	// Запускаем
	if a.Config.File {
		a.StartFile(wait, stop)
	}
	if a.Config.Stdout {
		a.StartStdout(wait, stop)
	}
	if a.Config.Mq {
		a.StartMQ(wait, stop)
	}
}

func (a *axAgent) StartFile(wait *sync.WaitGroup, stop chan bool) {
	a.Config.File = true

	if !a.file.started {
		a.file.SendFunc = a.file.Send // Заменяем "отправляющие" функции
		a.file.started = true
		go a.file.WriteFunc(wait, stop)
	}
}

func (a *axAgent) StartStdout(wait *sync.WaitGroup, stop chan bool) {
	a.Config.Stdout = true

	if !a.stdout.started {
		a.stdout.SendFunc = a.stdout.Send // Заменяем "отправляющие" функции
		a.stdout.started = true
		go a.stdout.WriteFunc(wait, stop)
	}
}

func (a *axAgent) StartMQ(wait *sync.WaitGroup, stop chan bool) {
	a.Config.Mq = true

	if len(a.mq.paths) == 0 {
		log.Print("Don't start mq because of empty configuration!")
	} else {
		if !a.mq.started {
			a.mq.SendFunc = a.mq.Send // Заменяем "отправляющие" функции
			if len(a.mq.paths) <= 0 {
				log.Fatal("Can't start mq logging option because of empty configuration!")
			} else {
				a.mq.started = true
				go a.mq.WriteFunc(wait, stop)
			}
		}
	}
}

// Добавить лог
func Add(s Severity, params map[MessageParamPreset]interface{}) error {
	if Agent != nil {
		return Agent.Add(s, params)
	} else {
		log.Fatal("logging agent is null")
	}
	return fmt.Errorf("logging agent is null")
}

// Добавить лог с выбором куда его помещать
func AddTo(s Severity, params map[MessageParamPreset]interface{}, dest Destinations) error {
	if Agent != nil {
		return Agent.AddTo(s, params, dest)
	} else {
		log.Fatal("logging agent is null")
	}
	return fmt.Errorf("logging agent is null")
}

// Добавить сообщение для записи
func (a *axAgent) Add(s Severity, params map[MessageParamPreset]interface{}) error {
	dest := Destinations{}
	dest.Any()
	return a.AddTo(s, params, dest)
}

// Добавить сообщение для записи с выбором куда его помещать
func (a *axAgent) AddTo(s Severity, params map[MessageParamPreset]interface{}, dest Destinations) error {
	m := Message{
		Time: time.Now(),
		Cat:  s,
		Params: map[MessageParamPreset]interface{}{
			Subsystem:   a.Location.Subsystem,
			Application: a.Location.App,
			Host:        a.Location.Host,
			ProcessID:   a.Location.ProcessId,
		},
	}
	// Добавляем параметры и проверяем на корректность типов
	for k, v := range params {
		// Проверим на корректность параметра и его заполненность
		if _, is := messagePresetTags[k]; !is {
			return fmt.Errorf("unknown parameter: %v", k)
		}
		// Проверим на корректность типа параметра
		switch vt := v.(type) {
		case string:
			if messageProfilePresetsTypes[k] != MessageParamString {
				return fmt.Errorf("param %s has invalid type %v, must be string", messagePresetTags[k], vt)
			}
			// Записываем значения строк только в случае если они не пустые
			if len(v.(string)) > 0 {
				m.Params[k] = v
			}
		case int, int64:
			if messageProfilePresetsTypes[k] != MessageParamInt {
				return fmt.Errorf("param %s has invalid type %v, must be int", messagePresetTags[k], vt)
			}
			m.Params[k] = v
		default:
			log.Printf("param \"%s\" has invalid type %v", messagePresetTags[k], vt)
			return fmt.Errorf("param \"%s\" has invalid type %v", messagePresetTags[k], vt)
		}
	}

	if dest.File && a.Config.File {
		a.file.Add(m)
	}
	if dest.Stdout && a.Config.Stdout {
		a.stdout.Add(m)
	}
	if dest.Mq && a.Config.Mq {
		a.mq.Add(m)
	}

	return nil
}

func (a *axAgent) AddInfo(params map[MessageParamPreset]interface{}) error {
	return a.Add(Information, params)
}

func (a *axAgent) AddError(params map[MessageParamPreset]interface{}) error {
	return a.Add(Error, params)
}

func (a *axAgent) AddDebug(params map[MessageParamPreset]interface{}) error {
	return a.Add(Debug, params)
}

func (a *axAgent) AddWarning(params map[MessageParamPreset]interface{}) error {
	return a.Add(Warning, params)
}

func (a *axAgent) AddSimpleDebug(action string, description string) error {
	return a.Add(Debug, map[MessageParamPreset]interface{}{Action: action, Description: description})
}

func (a *axAgent) AddSimpleError(action string, description string) error {
	return a.Add(Error, map[MessageParamPreset]interface{}{Action: action, Description: description})
}

func (a *axAgent) AddSimpleInfo(action string, description string) error {
	return a.Add(Information, map[MessageParamPreset]interface{}{Action: action, Description: description})
}
