package handlers_test

import (
	"PrintServer/PrintServer/diagnostic"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

const (
	exportPath   = "/api/core/print/export"
	previewPath  = "/api/core/print/preview"
	validatePath = "/api/core/print/svc/validate-template"
)

var validHash = fmt.Sprintf("%x", sha256.Sum256([]byte("<h1>#Monitoring.Date#</h1>")))
var internalHeader = map[string]string{"SubToken": "handlers-test-key"}

func TestExportHTML(t *testing.T) {
	called := false
	fake := &fakeTransformer{render: func(_ context.Context, id int, subsystem string, data any) (string, string, error) {
		called = true
		if id != 42 || subsystem != "news" || !reflect.DeepEqual(data, map[string]any{"rundown": map[string]any{"title": "Вечерние новости"}}) {
			t.Errorf("unexpected RenderTemplate arguments: id=%d subsystem=%q data=%#v", id, subsystem, data)
		}
		return "Вечерние новости", "<h1>Готово</h1>", nil
	}}
	response := request(t, newRouter(fake), exportPath, `{"template_id":42,"subsystem":"news","data":{"rundown":{"title":"Вечерние новости"}}}`, nil)
	if !called {
		t.Error("RenderTemplate was not called")
	}
	if response.Code != http.StatusOK || response.Body.String() != "<h1>Готово</h1>" {
		t.Errorf("response: status=%d body=%q", response.Code, response.Body.String())
	}
	if response.Header().Get("Content-Type") != "text/html" {
		t.Errorf("Content-Type: %q", response.Header().Get("Content-Type"))
	}
	if response.Header().Get("Content-Length") != "21" {
		t.Errorf("Content-Length: %q", response.Header().Get("Content-Length"))
	}
	if response.Header().Get("X-File-Name") != url.QueryEscape("Вечерние новости.html") {
		t.Errorf("X-File-Name: %q", response.Header().Get("X-File-Name"))
	}
	if !strings.Contains(response.Header().Get("Content-Disposition"), "filename*=UTF-8''"+url.QueryEscape("Вечерние новости.html")) {
		t.Errorf("Content-Disposition: %q", response.Header().Get("Content-Disposition"))
	}
}

func TestExportDOCX(t *testing.T) {
	converted := false
	fake := &fakeTransformer{
		render: func(_ context.Context, id int, subsystem string, data any) (string, string, error) {
			if id != 42 || subsystem != "news" {
				t.Errorf("unexpected RenderTemplate arguments: %d %q", id, subsystem)
			}
			return "report", "<p>Готово</p>", nil
		},
		docx: func(content string) ([]byte, error) {
			converted = true
			if content != "<p>Готово</p>" {
				t.Errorf("DOCX input: %q", content)
			}
			return []byte{0x50, 0x4b, 0x03, 0x04}, nil
		},
	}
	response := request(t, newRouter(fake), exportPath, `{"template_id":42,"subsystem":"news","data":{},"format":"docx"}`, nil)
	if !converted {
		t.Error("DOCXFromTemplate was not called")
	}
	if response.Code != http.StatusOK || !reflect.DeepEqual(response.Body.Bytes(), []byte{0x50, 0x4b, 0x03, 0x04}) {
		t.Errorf("response: status=%d body=%x", response.Code, response.Body.Bytes())
	}
	if response.Header().Get("Content-Type") != "application/vnd.openxmlformats-officedocument.wordprocessingml.document" {
		t.Errorf("Content-Type: %q", response.Header().Get("Content-Type"))
	}
	if response.Header().Get("X-File-Name") != "report.docx" {
		t.Errorf("X-File-Name: %q", response.Header().Get("X-File-Name"))
	}
}

func TestPreviewHTML(t *testing.T) {
	called := false
	fake := &fakeTransformer{preview: func(_ context.Context, content, subsystem, templateType string) (string, string, error) {
		called = true
		if content != "<h1>#Monitoring.Date#</h1>" || subsystem != "news" || templateType != "monitoring" {
			t.Errorf("unexpected PreviewTemplate arguments: %q %q %q", content, subsystem, templateType)
		}
		return "<h1>Предпросмотр</h1>", "Предпросмотр", nil
	}}
	response := request(t, newRouter(fake), previewPath, `{"subsystem":"news","template_type":"monitoring","template":"<h1>#Monitoring.Date#</h1>"}`, nil)
	if !called {
		t.Error("PreviewTemplate was not called")
	}
	if response.Code != http.StatusOK || response.Body.String() != "<h1>Предпросмотр</h1>" {
		t.Errorf("response: status=%d body=%q", response.Code, response.Body.String())
	}
	if response.Header().Get("Content-Type") != "text/html" {
		t.Errorf("Content-Type: %q", response.Header().Get("Content-Type"))
	}
	if response.Header().Get("X-File-Name") != url.QueryEscape("Предпросмотр") {
		t.Errorf("X-File-Name: %q", response.Header().Get("X-File-Name"))
	}
}

func TestValidateTemplateSuccess(t *testing.T) {
	called := false
	fake := &fakeTransformer{validate: func(content, subsystem, templateType, hash string, data map[string]any) error {
		called = true
		if content != "<h1>#Monitoring.Date#</h1>" || subsystem != "news" || templateType != "monitoring" || hash != validHash || !reflect.DeepEqual(data, map[string]any{"Monitoring": map[string]any{"Date": "2026-09-08"}}) {
			t.Errorf("unexpected ValidateTemplate arguments: %q %q %q %q %#v", content, subsystem, templateType, hash, data)
		}
		return nil
	}}
	body := `{"template_id":42,"subsystem":"news","template_type":"monitoring","template":"<h1>#Monitoring.Date#</h1>","content_hash":"` + validHash + `","data":{"Monitoring":{"Date":"2026-09-08"}}}`
	response := request(t, newRouter(fake), validatePath, body, internalHeader)
	if !called {
		t.Error("ValidateTemplate was not called")
	}
	if response.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", response.Code)
	}
	if !strings.HasPrefix(response.Header().Get("Content-Type"), "application/json") {
		t.Errorf("Content-Type: got %q, want application/json", response.Header().Get("Content-Type"))
	}
	var payload struct {
		Valid       bool   `json:"valid"`
		ContentHash string `json:"content_hash"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Errorf("response is not JSON: %v; body: %q", err, response.Body.String())
	} else if !payload.Valid || payload.ContentHash != validHash {
		t.Errorf("validation result: %+v", payload)
	}
}

func TestEndpointTemplateErrors(t *testing.T) {
	templateErr := diagnostic.NewTemplateError([]diagnostic.Diagnostic{{
		Code: diagnostic.CodeInvalidPath, Stage: diagnostic.StageTranspile,
		Message: "Неверный путь", Hint: "Исправьте путь",
		Span: &diagnostic.Span{Start: diagnostic.Location{Line: 2, Column: 3}, End: diagnostic.Location{Line: 2, Column: 8}},
	}}, nil)
	tests := []struct {
		name, path, body string
		fake             *fakeTransformer
		headers          map[string]string
	}{
		{"export", exportPath, `{"template_id":42,"subsystem":"news","data":{}}`, &fakeTransformer{render: func(context.Context, int, string, any) (string, string, error) { return "", "", templateErr }}, nil},
		{"preview", previewPath, `{"subsystem":"news","template_type":"monitoring","template":"#Bad..Path#"}`, &fakeTransformer{preview: func(context.Context, string, string, string) (string, string, error) { return "", "", templateErr }}, nil},
		{"validate", validatePath, `{"template_id":42,"subsystem":"news","template_type":"monitoring","template":"#Bad..Path#","content_hash":"` + validHash + `","data":{}}`, &fakeTransformer{validate: func(string, string, string, string, map[string]any) error { return templateErr }}, internalHeader},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := request(t, newRouter(tt.fake), tt.path, tt.body, tt.headers)
			requireError(t, response, http.StatusUnprocessableEntity, string(diagnostic.CodeInvalidPath))
			var payload struct {
				Fields  map[string][]string `json:"fields"`
				Details struct {
					Stage string           `json:"stage"`
					Span  *diagnostic.Span `json:"span"`
				} `json:"details"`
			}
			if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
				t.Fatal(err)
			}
			if payload.Details.Stage != string(diagnostic.StageTranspile) || payload.Details.Span == nil || payload.Details.Span.Start.Line != 2 || len(payload.Fields["template"]) != 1 {
				t.Errorf("template diagnostic details: %+v", payload)
			}
		})
	}
}

func TestExportConversionError(t *testing.T) {
	fake := &fakeTransformer{
		render: func(context.Context, int, string, any) (string, string, error) { return "report", "<p>x</p>", nil },
		docx: func(string) ([]byte, error) {
			return nil, diagnostic.NewTemplateError([]diagnostic.Diagnostic{{Code: diagnostic.CodeTemplateExecutionFailed, Stage: diagnostic.StageExecute, Message: "conversion failed"}}, errors.New("conversion failed"))
		},
	}
	response := request(t, newRouter(fake), exportPath, `{"template_id":42,"subsystem":"news","data":{},"format":"docx"}`, nil)
	requireError(t, response, http.StatusUnprocessableEntity, string(diagnostic.CodeTemplateExecutionFailed))
}

func TestMalformedJSON(t *testing.T) {
	for _, tt := range []struct {
		name, path, body string
		headers          map[string]string
	}{
		{"export", exportPath, `{`, nil},
		{"preview", previewPath, `{`, nil},
		{"validate", validatePath, `{`, internalHeader},
		{"trailing JSON", exportPath, `{"template_id":42,"subsystem":"news","data":{}} {}`, nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeTransformer{}
			response := request(t, newRouter(fake), tt.path, tt.body, tt.headers)
			requireError(t, response, http.StatusBadRequest, "INVALID_JSON")
			if len(fake.unexpected) != 0 {
				t.Errorf("transformer was called for malformed JSON: %v", fake.unexpected)
			}
		})
	}
}

func TestInvalidRequests(t *testing.T) {
	tests := []struct {
		name, path, body string
		headers          map[string]string
		status           int
		code             string
	}{
		{"export missing id", exportPath, `{"subsystem":"news","data":{}}`, nil, 422, "INVALID_REQUEST"},
		{"export missing data", exportPath, `{"template_id":42,"subsystem":"news"}`, nil, 422, "INVALID_REQUEST"},
		{"export bad subsystem", exportPath, `{"template_id":42,"subsystem":"Bad Name","data":{}}`, nil, 422, "INVALID_REQUEST"},
		{"export unsupported format", exportPath, `{"template_id":42,"subsystem":"news","data":{},"format":"pdf"}`, nil, 422, "UNSUPPORTED_FORMAT"},
		{"preview missing type", previewPath, `{"subsystem":"news","template":"hello"}`, nil, 422, "INVALID_REQUEST"},
		{"validate missing template id", validatePath, `{"subsystem":"news","template_type":"monitoring","template":"hello","content_hash":"` + validHash + `","data":{}}`, internalHeader, 422, "INVALID_REQUEST"},
		{"validate missing hash", validatePath, `{"template_id":42,"subsystem":"news","template_type":"monitoring","template":"hello","data":{}}`, internalHeader, 422, "INVALID_REQUEST"},
		{"validate missing data", validatePath, `{"template_id":42,"subsystem":"news","template_type":"monitoring","template":"hello","content_hash":"` + validHash + `"}`, internalHeader, 422, "INVALID_REQUEST"},
		{"validate zero template id", validatePath, `{"template_id":0,"subsystem":"news","template_type":"monitoring","template":"hello","content_hash":"` + validHash + `","data":{}}`, internalHeader, 422, "INVALID_REQUEST"},
		{"validate negative template id", validatePath, `{"template_id":-1,"subsystem":"news","template_type":"monitoring","template":"hello","content_hash":"` + validHash + `","data":{}}`, internalHeader, 422, "INVALID_REQUEST"},
		{"validate null template id", validatePath, `{"template_id":null,"subsystem":"news","template_type":"monitoring","template":"hello","content_hash":"` + validHash + `","data":{}}`, internalHeader, 422, "INVALID_REQUEST"},
		{"export wrong content type", exportPath, `{"template_id":42,"subsystem":"news","data":{}}`, map[string]string{"Content-Type": "text/plain"}, 415, "UNSUPPORTED_MEDIA_TYPE"},
		{"preview wrong content type", previewPath, `{"subsystem":"news","template_type":"monitoring","template":"hello"}`, map[string]string{"Content-Type": "text/plain"}, 415, "UNSUPPORTED_MEDIA_TYPE"},
		{"validate wrong content type", validatePath, `{"template_id":42,"subsystem":"news","template_type":"monitoring","template":"hello","content_hash":"` + validHash + `","data":{}}`, map[string]string{"Content-Type": "text/plain", "SubToken": "handlers-test-key"}, 415, "UNSUPPORTED_MEDIA_TYPE"},
		{"export unknown property", exportPath, `{"template_id":42,"subsystem":"news","data":{},"unexpected":true}`, nil, 422, "INVALID_REQUEST"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeTransformer{}
			response := request(t, newRouter(fake), tt.path, tt.body, tt.headers)
			requireError(t, response, tt.status, tt.code)
			if len(fake.unexpected) != 0 {
				t.Errorf("transformer was called for invalid request: %v", fake.unexpected)
			}
		})
	}
}

func TestInternalEndpointRequiresSubToken(t *testing.T) {
	response := request(t, newRouter(&fakeTransformer{}), validatePath, `{}`, nil)
	requireError(t, response, http.StatusUnauthorized, "UNAUTHORIZED")
}

func TestEndpointsRejectGET(t *testing.T) {
	router := newRouter(&fakeTransformer{})
	for _, path := range []string{exportPath, previewPath, validatePath} {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			if path == validatePath {
				req.Header.Set("SubToken", "handlers-test-key")
			}
			response := httptest.NewRecorder()
			router.ServeHTTP(response, req)
			if response.Code < http.StatusBadRequest {
				t.Errorf("GET %s: status %d, want an error status", path, response.Code)
			}
		})
	}
}
