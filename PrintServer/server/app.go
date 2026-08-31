package server

import (
	"PrintServer/PrintServer/handlers"
	"PrintServer/PrintServer/transformer"
	logging "PrintServer/agent"
	"context"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type App struct {
	Router *mux.Router
	Logger *log.Logger
	Server *http.Server
}

func (app *App) InitializeRouter(transformer_ transformer.Transformer) {
	export_handler := handlers.NewExportHandler(transformer_)

	app.Router.HandleFunc("/export", export_handler.TransformTemplate).Methods("POST")
}

func (app *App) Init(transformer_ transformer.Transformer) {
	mainRouter := mux.NewRouter()
	mainRouter.Use(LoggingMiddleware)
	app.Router = mainRouter.PathPrefix("/api/core/print").Subrouter()
	app.Router.Use(handlersMiddlewareJWTRSA)
	app.InitializeRouter(transformer_)
}

func (app *App) Run(address string) error {
	app.Server = &http.Server{
		Addr:    address,
		Handler: app.Router,
	}
	err := app.Server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		logging.Agent.AddSimpleError("Ошибка HTTP сервера", "")
		return err
	}

	return nil

}

func (app *App) Shutdown(shutdownCtx context.Context) error {
	if app.Server == nil {
		return nil
	}
	logging.Agent.AddSimpleInfo("HTTP сервер", "Завершение работы активных подключений...")
	return app.Server.Shutdown(shutdownCtx)
}
