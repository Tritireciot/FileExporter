package handlers_test

import (
	"PrintServer/PrintServer/diagnostic"
	"PrintServer/PrintServer/transformer"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"testing"
)

// Здесь ошибки трансформера имитируют ответы клиента ConfigAppServer из requests.go.
// /printcontent не участвует: этот запрос относится к другому источнику данных.
func TestEndpointLookupAndUpstreamStatuses(t *testing.T) {
	tests := []struct {
		name, path, body string
		upstreamError    error
		status           int
		code             string
		retryable        bool
	}{
		{
			name: "export template not found", path: exportPath,
			body:          `{"template_id":42,"subsystem":"news","data":{}}`,
			upstreamError: &transformer.TemplateFetchError{Cause: &transformer.UpstreamStatusError{StatusCode: http.StatusNotFound}},
			status:        http.StatusNotFound, code: "TEMPLATE_NOT_FOUND",
		},
		{
			name: "preview template type not found", path: previewPath,
			body:          `{"subsystem":"news","template_type":"monitoring","template":"hello"}`,
			upstreamError: &transformer.TemplateTypeNotFoundError{},
			status:        http.StatusNotFound, code: "TEMPLATE_TYPE_NOT_FOUND",
		},
		{
			name: "export template not published", path: exportPath,
			body:          `{"template_id":42,"subsystem":"news","data":{}}`,
			upstreamError: &transformer.TemplateFetchError{Cause: &transformer.UpstreamStatusError{StatusCode: http.StatusConflict}},
			status:        http.StatusConflict, code: "TEMPLATE_NOT_ACTIVE",
		},
		{
			name: "export ConfigAppServer unavailable", path: exportPath,
			body:          `{"template_id":42,"subsystem":"news","data":{}}`,
			upstreamError: &transformer.TemplateFetchError{Cause: &net.OpError{Op: "dial", Net: "tcp", Err: errors.New("connection refused")}},
			status:        http.StatusBadGateway, code: "CONFIG_SERVER_UNAVAILABLE", retryable: true,
		},
		{
			name: "export ConfigAppServer bad status", path: exportPath,
			body:          `{"template_id":42,"subsystem":"news","data":{}}`,
			upstreamError: &transformer.TemplateFetchError{Cause: &transformer.UpstreamStatusError{StatusCode: http.StatusInternalServerError}},
			status:        http.StatusBadGateway, code: "TEMPLATE_FETCH_FAILED", retryable: true,
		},
		{
			name: "preview ConfigAppServer unavailable", path: previewPath,
			body:          `{"subsystem":"news","template_type":"monitoring","template":"hello"}`,
			upstreamError: &transformer.PreviewDataFetchError{Cause: &net.OpError{Op: "dial", Net: "tcp", Err: errors.New("connection refused")}},
			status:        http.StatusBadGateway, code: "CONFIG_SERVER_UNAVAILABLE", retryable: true,
		},
		{
			name: "preview ConfigAppServer invalid JSON", path: previewPath,
			body:          `{"subsystem":"news","template_type":"monitoring","template":"hello"}`,
			upstreamError: &transformer.PreviewDataFetchError{Cause: errors.New("некорректный JSON от ConfigAppServer")},
			status:        http.StatusBadGateway, code: "PREVIEW_DATA_FETCH_FAILED", retryable: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeTransformer{}
			if tt.path == exportPath {
				fake.render = func(context.Context, int, string, any) (string, string, error) {
					return "", "", tt.upstreamError
				}
			} else {
				fake.preview = func(context.Context, string, string, string) (string, string, error) {
					return "", "", tt.upstreamError
				}
			}
			response := request(t, newRouter(fake), tt.path, tt.body, nil)
			requireError(t, response, tt.status, tt.code)
			var payload struct {
				Details struct {
					Stage     string `json:"stage"`
					Retryable bool   `json:"retryable"`
				} `json:"details"`
			}
			if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
				t.Fatalf("decode details: %v", err)
			}
			if payload.Details.Stage != "template_lookup" || payload.Details.Retryable != tt.retryable {
				t.Errorf("details: got %+v, want stage=template_lookup retryable=%v", payload.Details, tt.retryable)
			}
		})
	}
}

func TestEndpointUnexpectedError(t *testing.T) {
	tests := []struct {
		name, path, body string
		headers          map[string]string
	}{
		{"export", exportPath, `{"template_id":42,"subsystem":"news","data":{}}`, nil},
		{"preview", previewPath, `{"subsystem":"news","template_type":"monitoring","template":"hello"}`, nil},
		{"validate", validatePath, `{"template_id":42,"subsystem":"news","template_type":"monitoring","template":"hello","content_hash":"` + validHash + `","data":{}}`, internalHeader},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			internal := errors.New("private internal failure")
			fake := &fakeTransformer{
				render:   func(context.Context, int, string, any) (string, string, error) { return "", "", internal },
				preview:  func(context.Context, string, string, string) (string, string, error) { return "", "", internal },
				validate: func(string, string, string, string, map[string]any) error { return internal },
			}
			response := request(t, newRouter(fake), tt.path, tt.body, tt.headers)
			requireError(t, response, http.StatusInternalServerError, "INTERNAL_ERROR")
			if strings.Contains(response.Body.String(), internal.Error()) {
				t.Error("internal error leaked to client")
			}
		})
	}
}

func TestEndpointTemporarilyUnavailable(t *testing.T) {
	tests := []struct {
		name, path, body string
		headers          map[string]string
	}{
		{"export", exportPath, `{"template_id":42,"subsystem":"news","data":{}}`, nil},
		{"preview", previewPath, `{"subsystem":"news","template_type":"monitoring","template":"hello"}`, nil},
		{"validate", validatePath, `{"template_id":42,"subsystem":"news","template_type":"monitoring","template":"hello","content_hash":"` + validHash + `","data":{}}`, internalHeader},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unavailable := diagnostic.NewRequestError(http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "временно недоступен")
			fake := &fakeTransformer{
				render:   func(context.Context, int, string, any) (string, string, error) { return "", "", unavailable },
				preview:  func(context.Context, string, string, string) (string, string, error) { return "", "", unavailable },
				validate: func(string, string, string, string, map[string]any) error { return unavailable },
			}
			response := request(t, newRouter(fake), tt.path, tt.body, tt.headers)
			requireError(t, response, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE")
			var payload struct {
				Details struct {
					Retryable bool `json:"retryable"`
				} `json:"details"`
			}
			if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
				t.Fatalf("decode details: %v", err)
			}
			if !payload.Details.Retryable {
				t.Error("503 response must be retryable")
			}
		})
	}
}

func TestPrintContentFailureIsNotConfigAppServerFailure(t *testing.T) {
	fake := &fakeTransformer{render: func(context.Context, int, string, any) (string, string, error) {
		return "", "", fmt.Errorf("Получение данных шаблона: %w", &net.OpError{Op: "dial", Net: "tcp", Err: errors.New("connection refused")})
	}}
	response := request(t, newRouter(fake), exportPath, `{"template_id":42,"subsystem":"news","data":{}}`, nil)
	requireError(t, response, http.StatusInternalServerError, "INTERNAL_ERROR")
}
