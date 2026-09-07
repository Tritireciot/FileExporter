package config

import (
	"AutoplayX/deploy"
	"AutoplayX/pgutils"
	"fmt"
	"strconv"

	"gopkg.in/ini.v1"
)

func GetIniConfiguration(fileName string) error {
	cfg, err := ini.Load(fileName)
	if err != nil {
		return fmt.Errorf("fail to read file: %v", err)
	} else {
		{ // SubAccess
			sa := deploy.AccessConfig{}

			sa.Path = cfg.Section("SubAccess").Key("path").String()
			sa.Port, _ = strconv.Atoi(cfg.Section("SubAccess").Key("port").String())

			SubAccess = append(SubAccess, sa)
		}
		//-----------------------------------------
		{ // Database
			if cfg.HasSection("Database") {
				dbc := pgutils.DatabaseConfig{}
				dbc.Port, err = cfg.Section("Database").Key("serverPort").Int()
				if err != nil {
					dbc.Port = 5432
					dbc.User = cfg.Section("Database").Key("userName").String()
				}
				dbc.Password = cfg.Section("Database").Key("password").String()
				dbc.DBname = cfg.Section("Database").Key("dbName").String()
				dbc.Host = cfg.Section("Database").Key("serverName").String()

				DbConfig = append(DbConfig, dbc)
			}

		}
		{ // SubAccess
			mq := deploy.AccessConfig{}

			mq.Path = cfg.Section("MessageQueue").Key("path").String()
			mq.Port, _ = strconv.Atoi(cfg.Section("MessageQueue").Key("port").String())

			MQAccess = append(MQAccess, mq)
		}

		// etc
		MessAppSrv.HttpPort, err = cfg.Section("App").Key("httpPort").Int()
		if err != nil {
			MessAppSrv.HttpPort = DefaultHttpPort
		}

		return nil
	}
}
