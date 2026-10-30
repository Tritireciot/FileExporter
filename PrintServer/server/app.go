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

func (app *App) InitializeRouter(transformer_ transformer.Transformer, publicRouter *mux.Router, internalRouter *mux.Router) {
	export_handler := handlers.NewExportHandler(transformer_)

	publicRouter.HandleFunc("/export", export_handler.TransformTemplate).Methods("POST")
	publicRouter.HandleFunc("/preview", export_handler.PreviewTemplate).Methods("POST")

	internalRouter.HandleFunc("/validate-template", export_handler.ValidateTemplate).Methods("POST")
}

func (app *App) Init(transformer_ transformer.Transformer) {
	app.Router = mux.NewRouter()
	app.Router.Use(LoggingMiddleware)
	publicRouter := app.Router.PathPrefix("/api/core/print").Subrouter()
	//app.Router.Use(handlersMiddlewareJWTRSA)

	internalRouter := app.Router.PathPrefix("/api/core/print/svc").Subrouter()
	internalRouter.Use(handlersInternalMiddleware)
	app.InitializeRouter(transformer_, publicRouter, internalRouter)
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
