package handlers_test

import (
	"PrintServer/PrintServer/config"
	"PrintServer/PrintServer/server"
	"PrintServer/PrintServer/transformer"
	logging "PrintServer/agent"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeTransformer struct {
	render     func(context.Context, int, string, any) (string, string, error)
	preview    func(context.Context, string, string, string) (string, string, error)
	validate   func(string, string, string, string, map[string]any) error
	docx       func(string) ([]byte, error)
	unexpected []string
}

var _ transformer.Transformer = (*fakeTransformer)(nil)

func (f *fakeTransformer) RenderTemplate(ctx context.Context, id int, subsystem string, data any) (string, string, error) {
	if f.render == nil {
		f.unexpected = append(f.unexpected, "RenderTemplate")
		return "", "", nil
	}
	return f.render(ctx, id, subsystem, data)
}

func (f *fakeTransformer) PreviewTemplate(ctx context.Context, content, subsystem, templateType string) (string, string, error) {
	if f.preview == nil {
		f.unexpected = append(f.unexpected, "PreviewTemplate")
		return "", "", nil
	}
	return f.preview(ctx, content, subsystem, templateType)
}

func (f *fakeTransformer) ValidateTemplate(content, subsystem, templateType, hash string, data map[string]any) error {
	if f.validate == nil {
		f.unexpected = append(f.unexpected, "ValidateTemplate")
		return nil
	}
	return f.validate(content, subsystem, templateType, hash, data)
}

func (f *fakeTransformer) DOCXFromTemplate(content string) ([]byte, error) {
	if f.docx == nil {
		f.unexpected = append(f.unexpected, "DOCXFromTemplate")
		return nil, nil
	}
	return f.docx(content)
}

func (f *fakeTransformer) PDFFromTemplate(string) ([]byte, error) {
	panic("PDF is not part of the published API")
}

func TestMain(m *testing.M) {
	logging.InitAXAgent(logging.Config{}, logging.AXLocation{})
	m.Run()
}

func newRouter(fake *fakeTransformer) http.Handler {
	config.SubAccessCommKey = "handlers-test-key"
	app := &server.App{}
	app.Init(fake)
	return app.Router
}

func request(t *testing.T, router http.Handler, path, body string, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	for name, value := range headers {
		req.Header.Set(name, value)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	return response
}

func requireError(t *testing.T, response *httptest.ResponseRecorder, status int, code string) {
	t.Helper()
	if response.Code != status {
		t.Errorf("status: got %d, want %d; body: %s", response.Code, status, response.Body.String())
	}
	if !strings.HasPrefix(response.Header().Get("Content-Type"), "application/json") {
		t.Errorf("Content-Type: got %q, want application/json", response.Header().Get("Content-Type"))
		return
	}
	var payload struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Errorf("response is not JSON: %v; body: %q", err, response.Body.String())
		return
	}
	if payload.Code != code {
		t.Errorf("code: got %q, want %q", payload.Code, code)
	}
	if payload.Message == "" {
		t.Error("error message is empty")
	}
}
