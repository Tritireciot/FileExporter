package main

import (
	"PrintServer/PrintServer/config"
	"PrintServer/PrintServer/db"
	"PrintServer/PrintServer/server"
	"PrintServer/PrintServer/transformer"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
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

	defer func() {
		close(stopLogger)
		waitLogger.Wait()
	}()

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
				return
			}
			break
		}
		time.Sleep(time.Second * 5)
	}

	if db_repo != nil {
		defer db_repo.TearDown()
	}

	transformer_ := transformer.NewTransformService(db_repo)
	app.Init(db_repo, transformer_)
	go func() {
		logging.Agent.AddSimpleInfo("HTTP сервер", "Запуск на порту: "+app_config.App.Port)
		if err := app.Run(app_config.ConfigureAppUrl()); err != nil && err != http.ErrServerClosed {
			logging.Agent.AddSimpleError("HTTP сервер", "Ошибка: "+err.Error())
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	logging.Agent.AddSimpleInfo("Системный сигнал", fmt.Sprintf("Получен сигнал %s. Начинаем плавную остановку...", sig.String()))
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := app.Shutdown(shutdownCtx); err != nil {
		logging.Agent.AddSimpleError("HTTP сервер", "Ошибка при плавном закрытии: "+err.Error())
	}

	logging.Agent.AddSimpleInfo("Статус", "Сервис успешно остановлен.")
}
