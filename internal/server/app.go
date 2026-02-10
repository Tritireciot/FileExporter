package server

import (
	"former/internal/db"
	"former/internal/handlers"
	"former/internal/transformer"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type App struct {
	Router *mux.Router
	Logger *log.Logger
}

func (app *App) InitializeRouter(db_repo db.DBRepo, transformer_ *transformer.TransformService) {
	template_handler := handlers.NewTemplateHandler(db_repo)
	tag_handler := handlers.NewTagsHandler(db_repo)
	export_handler := handlers.NewExportHandler(transformer_)
	app.Router.HandleFunc("/db/get_all_templates", template_handler.GetAllTemplates).Methods("GET")
	app.Router.HandleFunc("/db/get_template", template_handler.GetTemplate).Methods("GET")
	app.Router.HandleFunc("/db/add_template", template_handler.AddNewTemplate).Methods("POST")
	app.Router.HandleFunc("/db/delete_template", template_handler.DeleteTemplate).Methods("DELETE")

	app.Router.HandleFunc("/db/get_all_tags", tag_handler.GetAllTags).Methods("GET")
	app.Router.HandleFunc("/db/add_tag", tag_handler.AddNewTag).Methods("POST")
	app.Router.HandleFunc("/db/delete_tag", tag_handler.DeleteTag).Methods("DELETE")

	app.Router.HandleFunc("/export", export_handler.TransformTemplate).Methods("POST")
}

func (app *App) Init(db_repo db.DBRepo, transformer_ *transformer.TransformService) {
	app.Router = mux.NewRouter()
	app.InitializeRouter(db_repo, transformer_)
}

func (app *App) Run(address string) {
	loggedRouter := app.createLoggingRouter(app.Logger.Writer())
	app.Logger.Fatal(http.ListenAndServe(address, loggedRouter))
}
