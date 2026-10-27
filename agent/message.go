package logging

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

type Severity int

const (
	Emergency   Severity = 1
	Alert       Severity = 2
	Critical    Severity = 3
	Error       Severity = 4
	Warning     Severity = 5
	Notice      Severity = 6
	Information Severity = 7
	Debug       Severity = 8
)

var severityTags = map[Severity]string{
	Emergency:   "EMERGENCY",
	Alert:       "ALERT",
	Critical:    "CRITICAL",
	Error:       "ERROR",
	Warning:     "WARNING",
	Notice:      "NOTICE",
	Information: "INFORMATION",
	Debug:       "DEBUG",
}

type MessageParamPreset int

const (
	Title       MessageParamPreset = 0
	Subsystem   MessageParamPreset = 1
	Application MessageParamPreset = 2
	Host        MessageParamPreset = 3
	ProcessID   MessageParamPreset = 4
	Action      MessageParamPreset = 5
	Description MessageParamPreset = 6
	User        MessageParamPreset = 7
	UserID      MessageParamPreset = 8
	ObjectType  MessageParamPreset = 9
	ObjectName  MessageParamPreset = 10
	ObjectID    MessageParamPreset = 11
	ClipID      MessageParamPreset = 12
	SourceIP    MessageParamPreset = 13
	SessionID   MessageParamPreset = 14
	// etc...
)

type MessageParamType string

const (
	MessageParamInt    MessageParamType = "int"
	MessageParamString MessageParamType = "string"
	MessageParamDate   MessageParamType = "date"
)

// Логируемый параметр сообщения
type MessageParamProps struct {
	Type  MessageParamType
	Value interface{}
}

var messagePresetTags = map[MessageParamPreset]string{
	Title:       "title",
	Subsystem:   "subsystem",
	Application: "app",
	Host:        "host",
	ProcessID:   "processId",
	Action:      "action",
	Description: "description",
	User:        "user",
	UserID:      "userId",
	ObjectType:  "objectType",
	ObjectName:  "objectName",
	ObjectID:    "objectId",
	ClipID:      "clipId",
	SourceIP:    "sourceIP",
	SessionID:   "sessionID",
}

var messageProfilePresetsTypes = map[MessageParamPreset]MessageParamType{
	Title:       MessageParamString,
	Subsystem:   MessageParamString,
	Application: MessageParamString,
	Host:        MessageParamString,
	ProcessID:   MessageParamInt,
	Action:      MessageParamString,
	Description: MessageParamString,
	User:        MessageParamString,
	UserID:      MessageParamInt,
	ObjectType:  MessageParamString,
	ObjectName:  MessageParamString,
	ObjectID:    MessageParamInt,
	ClipID:      MessageParamInt,
	SourceIP:    MessageParamString,
	SessionID:   MessageParamInt,
}

type AXLocation struct {
	Subsystem string `json:"subsystem"`
	App       string `json:"app"`
	Host      string `json:"host"`
	ProcessId uint32 `json:"processId"`
}

// Базовая структура сообщения
type Message struct {
	Time time.Time
	Cat  Severity

	// Параметры
	Params map[MessageParamPreset]interface{}
}

func (m *Message) MarshalJSON() ([]byte, error) {

	jstr := map[string]interface{}{}

	jstr["time"] = m.Time
	jstr["category"] = m.Cat

	for k, v := range m.Params {
		jstr[messagePresetTags[k]] = v
	}

	b, err := json.Marshal(jstr)

	return b, err
}

func (m *Message) Text() string {

	str := fmt.Sprintf(
		`%s
DATE: %s
TIME: %s`,
		severityTags[m.Cat], m.Time.Format("2006-01-02"), m.Time.Format("15:04:05.000"))

	keys := []MessageParamPreset{}
	for k := range m.Params {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	for _, k := range keys {
		str += fmt.Sprintf("\n%s: %v", strings.ToUpper(messagePresetTags[k]), m.Params[k])
	}

	return str
}
