package main

import (
	"PrintServer/PrintServer/config"
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
		Subsystem: "Core",
		App:       "Print",
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
	for {
		if err := config.GetConfig(); err != nil {
			logging.Agent.AddSimpleError("Загрузка конфигурации", "Не удалось загрузить конфигурацию приложения: "+err.Error())
		} else {
			break
		}
		time.Sleep(time.Second * 5)
	}

	transformer_ := transformer.NewTransformService()
	app.Init(transformer_)
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
