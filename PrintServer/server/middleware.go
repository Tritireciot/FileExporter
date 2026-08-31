package server

import (
	"AutoplayX/jwtoken"
	"PrintServer/PrintServer/config"
	logging "PrintServer/agent"
	"bytes"
	"context"
	"fmt"
	"net/http"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
	MaxLogLen  int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (w *responseWriter) Write(b []byte) (int, error) {
	if w.body.Len() < w.MaxLogLen {
		available := w.MaxLogLen - w.body.Len()
		if len(b) < available {
			w.body.Write(b)
		} else {
			w.body.Write(b[:available])
			w.body.WriteString("...")
		}
	}
	return w.ResponseWriter.Write(b)
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

		rw := &responseWriter{
			ResponseWriter: writer,
			statusCode:     http.StatusOK,
			body:           bytes.NewBufferString(""),
			MaxLogLen:      100,
		}

		endpointRequestStr := fmt.Sprintf(
			"Endpoint Method: %s %s",
			request.Method,
			request.URL.Path,
		)

		queryParams := request.URL.Query()

		if len(queryParams) > 0 {
			endpointRequestStr += "?" + queryParams.Encode()
		}

		next.ServeHTTP(rw, request)

		logging.Agent.AddSimpleInfo(
			endpointRequestStr,
			fmt.Sprintf("Status: %d Response: %s",
				rw.statusCode,
				rw.body.String(),
			),
		)

	})
}

func GetReqUserClaims(req *http.Request) (jwtoken.UserClaims, error) {
	tokenHeader := ""
	if c, err := req.Cookie("Authorization"); err == nil {
		tokenHeader = c.Value
	} else {
		tokenHeader = req.Header.Get("Authorization")
	}

	return jwtoken.GetUserClaims(&config.PublicKey, jwtoken.ParseHeaderToken(tokenHeader))
}

func handlersMiddlewareJWTRSA(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверка на наличие сетевого ключа
		// Открываем токен
		uclaims, err := GetReqUserClaims(r)
		if err != nil || !uclaims.Valid() {
			code := http.StatusUnauthorized
			http.Error(w, http.StatusText(code), code)
			return
		}
		// Добавляем и рассылаем всем
		ctx := context.WithValue(r.Context(), "uclaims", uclaims)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
