package config

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v4/pgxpool"
)

type NotificationEvent struct {
	ID          int              `json:"id"`
	Name        LangMapString    `json:"name"`
	Description LangMapString    `json:"description"`
	DeepChilds  []DeepChildGroup `json:"deep_childs,omitempty"`
	Push        bool             `json:"push,omitempty"`
	Email       bool             `json:"email,omitempty"`
	Max         bool             `json:"max,omitempty"`
	Tg          bool             `json:"tg,omitempty"`
}

type DeepChildGroup struct {
	Name  string     `json:"name"`
	Types []DeepType `json:"types"`
}

type DeepType struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Value bool   `json:"value"`
}

type NotificationGroup struct {
	Name   LangMapString       `json:"name"`
	Childs []NotificationEvent `json:"childs"`
}

type NotificationEventsSchema struct {
	NotificationEvents  []NotificationGroup `json:"NotificationEvents,omitempty"`
	NotificationEvents2 []NotificationGroup `json:"notificationEvents,omitempty"`
	NotificationEvents3 []NotificationGroup `json:"notificationEventsConfig,omitempty"`
}

func GetNotificationEventsSchema(pool *pgxpool.Pool) (NotificationEventsSchema, error) {
	schema := NotificationEventsSchema{}

	if pool == nil {
		return schema, nil
	}

	conn, err := pool.Acquire(context.Background())
	if err != nil {
		return schema, err
	}
	defer conn.Release()

	val, _, err := GetSQLConfigValue(conn.Conn(), "notificationEvents.schema")
	if err != nil {
		return schema, err
	}

	err = json.Unmarshal([]byte(val), &schema)
	return schema, err
}

func GetDefaultNotificationEvents(pool *pgxpool.Pool) ([]NotificationEvent, error) {
	schema, err := GetNotificationEventsSchema(pool)
	if err != nil {
		return nil, err
	}

	var events []NotificationEvent
	for _, group := range schema.NotificationEvents {
		events = append(events, group.Childs...)
	}

	return events, nil
}
