package config

import (
	"AutoplayX/deploy"
	"AutoplayX/pgutils"
	"crypto/rsa"
	"strconv"
)

var (
	SubAccess     []deploy.AccessConfig
	MQAccess      []deploy.AccessConfig // Пути к брокерам очереди сообщений
	SearchAddress []deploy.AccessConfig
	DbConfig      []pgutils.DatabaseConfig

	PublicKey        rsa.PublicKey
	SubAccessCommKey string

	SubAccessToken string
)

// --------------------------------------------------------------------------------------
// Конфигурация сервиса при получении через GetServicesByType
type serviceConfiguration struct {
	Path          string      `json:"path"`
	Configuration interface{} `json:"configuration"`
}

var (
	ServiceID string // ИД службы
)

func GetServiceID() int {
	id, _ := strconv.Atoi(ServiceID)
	return id
}
