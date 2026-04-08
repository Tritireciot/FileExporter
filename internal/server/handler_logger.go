package server

import (
	"io"
	"net/http"

	"github.com/gorilla/handlers"
)

func (app *App) createLoggingRouter(out io.Writer) http.Handler {
	return handlers.LoggingHandler(out, app.Router)
}
