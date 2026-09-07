package config

import (
	configuration "AutoplayX/configuration"
	"AutoplayX/deploy"
	"AutoplayX/pgutils"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/alecthomas/kingpin/v2"
)

var (
	argServiceId = kingpin.Flag("sid", "Service id.").Short('s').String()
	argCfgMethod = kingpin.Flag("method", "Configuration get method (ini, psql).").Short('m').String()
	argCfgPath   = kingpin.Flag("path", "Configuration get path (ini - path to file, psql - CFG_PATH=plan:config:127.0.0.1:5432:postgres:postgres).").Short('p').String()
)

// --------------------------------------------------------------------------------------
func getServiceID() {
	// Сначала попробуем как env
	ServiceID = os.Getenv(ServiceIDTag)
	if len(ServiceID) <= 0 && argServiceId != nil { // Попробуем взять как параметр запуска
		ServiceID = *argServiceId
	}
	if len(ServiceID) <= 0 {
		log.Fatalf("No service id! Set %s env!", ServiceIDTag)
	}
}

// -------------------------------------------------------------------------------------
func getConfigurationMethod() (string, string) {
	// Источник получения конфигурации по-умолчанию - ini
	cfgMethod := deploy.CFG_INI_METHOD
	cfgPath := ""
	// Установим по-умолчанию название файла ini конфигурации
	if executable, err := os.Executable(); err == nil {
		cfgPath = PrepareConfigIniName(executable)
	}

	// Пробуем получить первоначальнные данные через переменне среды
	cfmenv := os.Getenv(deploy.CFG_METHOD_TAG)
	if len(cfmenv) > 0 {
		cfgMethod = os.Getenv(deploy.CFG_METHOD_TAG)
	} else if argCfgMethod != nil {
		cfgMethod = *argCfgMethod
	}

	cfmpath := os.Getenv(deploy.CFG_PATH_TAG)
	if len(cfmpath) > 0 {
		cfgPath = os.Getenv(deploy.CFG_PATH_TAG)
	} else if argCfgPath != nil {
		cfgPath = *argCfgPath
	}

	return cfgMethod, cfgPath
}

// --------------------------------------------------------------------------------------
func PrepareConfigIniName(executable string) string {
	return executable[:len(executable)-len(filepath.Ext(executable))] + ".ini"
}

// --------------------------------------------------------------------------------------
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

// --------------------------------------------------------------------------------------
func GetConfig() error {
	kingpin.Parse()
	//---------------------
	// Получим ИД службы
	getServiceID()
	cfgMethod, cfgPath := getConfigurationMethod()

	if strings.ToLower(cfgMethod) == deploy.CFG_INI_METHOD {
		return GetIniConfiguration(cfgPath)
	} else if strings.ToLower(cfgMethod) == deploy.CFG_PSQL_METHOD {
		var dbcfg pgutils.DatabaseConfig
		var err error
		if dbcfg, err = configuration.GetDBConfigFromPath(cfgPath); err == nil {
			var cfg configuration.InfrastructureConfiguration
			if cfg, err = configuration.GetPSQLCommonConfiguration(dbcfg); err == nil {
				SubAccess = cfg.SubAccess
				MQAccess = cfg.MQAccess
				DbConfig = cfg.DBConfig
				PublicKey = cfg.PublicKey
				SubAccessCommKey = cfg.IntenalCommunicationToken
				MessAppSrv.HttpPort = GetEvironmentPort(HttpPortTag, DefaultHttpPort)
				MessAppSrv.WsPort = GetEvironmentPort(WSPortTag, DefaultWSPort)
				SubAccessToken = cfg.IntenalCommunicationToken
			} else {
				// Если PSQL конфиг не загрузился, устанавливаем значения по умолчанию
				MessAppSrv.HttpPort = DefaultHttpPort
				MessAppSrv.WsPort = DefaultWSPort
			}
		}
		return err
	}
	return fmt.Errorf("not supported get config method: %s", cfgMethod)
}

// --------------------------------------------------------------------------------------
func GetDBConfig() *pgutils.DatabaseConfig {
	if len(DbConfig) < 1 {
		log.Fatal("No database configuration!")
	}
	return &DbConfig[0]
}

// --------------------------------------------------------------------------------------
func MakePathList(list []deploy.AccessConfig) []string {
	pathlist := []string{}
	for i := range list {
		pathlist = append(pathlist, list[i].GetPath())
	}
	return pathlist
}

// --------------------------------------------------------------------------------------
func GetMQPathList() []string {
	return MakePathList(MQAccess)
}

func GetSubAccess() deploy.AccessConfig {
	if len(SubAccess) < 1 {
		log.Fatal("No subaccess configuration!")
	}
	return SubAccess[0]
}
