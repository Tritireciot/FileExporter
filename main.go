package main

import (
	"PrintServer/internal/config"
	"PrintServer/internal/db"
	"PrintServer/internal/server"
	"PrintServer/internal/transformer"
	"context"
	"os"
	"sync"
	"time"

	logging "PrintServer/agent"
)

func main() {

	host, _ := os.Hostname()

	logging.InitAXAgent(logging.Config{}, logging.AXLocation{
		Subsystem: "Print",
		App:       "Service",
		Host:      host,
		ProcessId: uint32(os.Getpid()),
	})

	waitLogger := sync.WaitGroup{}
	stopLogger := make(chan bool)

	logging.Agent.StartStdout(&waitLogger, stopLogger)
	logging.Agent.StartFile(&waitLogger, stopLogger)

	app := server.App{}
	app_config := config.LoadConfig()
	var db_repo *db.DBRepository
	for {
		if err := config.GetConfig(); err != nil {
			logging.Agent.AddSimpleError("Загрузка конфигурации", "Не удалось загрузить конфигурацию приложения: "+err.Error())
		} else {
			db_config := config.GetDBConfig()
			if os.Getenv("LOCAL") == "true" {
				db_config.DBname = "backend-db"
				db_config.Host = "db"
				db_config.User = "user"
				db_config.Password = "pass"
				db_config.Schema = "print"
			}
			db_repo, err = db.SetupDB(context.Background(), db_config)
			if err != nil {
				logging.Agent.AddSimpleError("Загрузка конфигурации", "Не удалось загрузить конфигурацию приложения: "+err.Error())
				os.Exit(1)
			}
			break
		}
		time.Sleep(time.Second * 5)
	}
	transformer_ := transformer.NewTransformService(db_repo)
	app.Init(db_repo, transformer_)
	logging.Agent.AddSimpleInfo("HTTP сервер", "Запуск на порту: "+ app_config.App.Port)
	if err := app.Run(app_config.ConfigureAppUrl()); err != nil {
		logging.Agent.AddSimpleError("HTTP сервер", "Ошибка: "+err.Error())
	}
}
