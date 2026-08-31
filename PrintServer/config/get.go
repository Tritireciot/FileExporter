package config

import (
	extConfig "AutoplayX/configuration"
	"AutoplayX/deploy"
	"AutoplayX/pgutils"
	"crypto/rsa"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

const (
	DefaultHttpPort  = 8153
	DefaultMASWSPort = 8151

	ServiceIDTag = "SERVICE_ID"
	HttpPortTag  = "HTTP_PORT"
	MASWSPortTag = "MAS_WS_PORT"
)

var Ports struct {
	HTTPPort  int
	MASWSPort int
}

var (
	ServiceID string // ИД службы
	SubAccess []deploy.AccessConfig
	DbConfig  []pgutils.DatabaseConfig

	PublicKey        rsa.PublicKey
	SubAccessCommKey string
)

func getServiceID() {
	// Сначала попробуем как env
	ServiceID = os.Getenv(ServiceIDTag)

	if len(ServiceID) <= 0 {
		log.Fatalf("No service id! Set %s env!", ServiceIDTag)
	}
}

func getConfigurationMethod() (string, string) {
	// Источник получения конфигурации по-умолчанию - ini
	cfgMethod := ""
	cfgPath := ""

	// Пробуем получить первоначальнные данные через переменне среды
	cfmenv := os.Getenv(deploy.CFG_METHOD_TAG)
	if len(cfmenv) > 0 {
		cfgMethod = os.Getenv(deploy.CFG_METHOD_TAG)
	}

	cfmpath := os.Getenv(deploy.CFG_PATH_TAG)
	if len(cfmpath) > 0 {
		cfgPath = os.Getenv(deploy.CFG_PATH_TAG)
	}

	return cfgMethod, cfgPath
}

func GetEvironmentPort(envtag string, defaultPort int) int {
	envValue := os.Getenv(envtag)
	if envValue == "" {
		return defaultPort
	}

	if port, err := strconv.Atoi(envValue); err != nil {
		log.Print(err)
		return defaultPort
	} else {
		return port
	}
}

func GetSubAccess() deploy.AccessConfig {
	if len(SubAccess) < 1 {
		log.Fatal("No subaccess configuration!")
	}
	return SubAccess[0]
}

func GetDBConfig() *pgutils.DatabaseConfig {
	if len(DbConfig) < 1 {
		log.Fatal("No database configuration!")
	}
	return &DbConfig[0]
}

func GetConfig() error {
	getServiceID()
	cfgMethod, cfgPath := getConfigurationMethod()

	if strings.ToLower(cfgMethod) == deploy.CFG_PSQL_METHOD {
		var dbcfg pgutils.DatabaseConfig
		var err error
		if dbcfg, err = extConfig.GetDBConfigFromPath(cfgPath); err == nil {
			var cfg extConfig.InfrastructureConfiguration
			if cfg, err = extConfig.GetPSQLCommonConfiguration(dbcfg); err == nil {
				SubAccess = cfg.SubAccess

				DbConfig = cfg.DBConfig
				PublicKey = cfg.PublicKey
				SubAccessCommKey = cfg.IntenalCommunicationToken
				Ports.HTTPPort = GetEvironmentPort(HttpPortTag, DefaultHttpPort)
				Ports.MASWSPort = GetEvironmentPort(MASWSPortTag, DefaultMASWSPort)
			} else {
				// Если PSQL конфиг не загрузился, устанавливаем значения по умолчанию
				Ports.HTTPPort = DefaultHttpPort
				Ports.MASWSPort = DefaultMASWSPort
			}
		}
		return err

	}
	return fmt.Errorf("not supported get config method: %s", cfgMethod)
}
