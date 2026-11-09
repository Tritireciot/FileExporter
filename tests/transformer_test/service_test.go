package transformer_test

import (
	"AutoplayX/deploy"
	"PrintServer/PrintServer/config"
	"PrintServer/PrintServer/diagnostic"
	"PrintServer/PrintServer/transformer"
	logging "PrintServer/agent"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
)

func TestMain(m *testing.M) {
	logging.InitAXAgent(logging.Config{}, logging.AXLocation{})
	m.Run()
}

func hash(content string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(content)))
}

func hostPort(t *testing.T, address string) (string, int) {
	t.Helper()
	host, portText, err := net.SplitHostPort(strings.TrimPrefix(address, "http://"))
	if err != nil {
		t.Fatal(err)
	}
	port, err := strconv.Atoi(portText)
	if err != nil {
		t.Fatal(err)
	}
	return host, port
}

func newService(t *testing.T, configHandler, printHandler http.HandlerFunc) *transformer.TransformService {
	t.Helper()
	configServer := httptest.NewServer(configHandler)
	printServer := httptest.NewServer(printHandler)
	t.Cleanup(configServer.Close)
	t.Cleanup(printServer.Close)

	configHost, configPort := hostPort(t, configServer.URL)
	printHost, printPort := hostPort(t, printServer.URL)
	oldSubAccess := config.SubAccess
	oldPrintHost, oldPrintPort := config.PrintContentHost, config.PrintContentPort
	oldEntries, oldBytes := config.PrintCacheMainMaxEntries, config.PrintCacheMainMaxBytes
	config.SubAccess = []deploy.AccessConfig{{Path: configHost, Port: configPort}}
	config.PrintContentHost, config.PrintContentPort = printHost, strconv.Itoa(printPort)
	config.PrintCacheMainMaxEntries, config.PrintCacheMainMaxBytes = 8, 1<<20
	t.Cleanup(func() {
		config.SubAccess = oldSubAccess
		config.PrintContentHost, config.PrintContentPort = oldPrintHost, oldPrintPort
		config.PrintCacheMainMaxEntries, config.PrintCacheMainMaxBytes = oldEntries, oldBytes
	})
	return transformer.NewTransformService()
}

func writeJSON(t *testing.T, w http.ResponseWriter, value any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		t.Errorf("write JSON: %v", err)
	}
}

func TestRenderTemplateUsesCurrentPublishedRevision(t *testing.T) {
	var mu sync.RWMutex
	content := "#Value#"
	configCalls, printCalls := 0, 0
	service := newService(t,
		func(w http.ResponseWriter, r *http.Request) {
			mu.Lock()
			defer mu.Unlock()
			configCalls++
			if r.Method != http.MethodGet || r.URL.Path != "/api/news/config/svc/print/templates/42/published" {
				t.Errorf("unexpected ConfigAppServer request: %s %s", r.Method, r.URL.Path)
			}
			writeJSON(t, w, map[string]any{"name": "report", "template_type": "rundown", "content": content, "content_hash": hash(content)})
		},
		func(w http.ResponseWriter, r *http.Request) {
			mu.Lock()
			defer mu.Unlock()
			printCalls++
			if r.Method != http.MethodPost || r.URL.Path != "/api/news/svc/printcontent" {
				t.Errorf("unexpected printcontent request: %s %s", r.Method, r.URL.Path)
			}
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body["type"] != "rundown" {
				t.Errorf("unexpected printcontent body: %#v, %v", body, err)
			}
			writeJSON(t, w, map[string]any{"Value": "ok"})
		},
	)

	name, output, err := service.RenderTemplate(context.Background(), 42, "news", map[string]any{"id": 1})
	if err != nil || name != "report" || output != "ok" {
		t.Fatalf("first render: name=%q output=%q err=%v", name, output, err)
	}
	mu.Lock()
	content = "new:#Value#"
	mu.Unlock()
	name, output, err = service.RenderTemplate(context.Background(), 42, "news", map[string]any{"id": 2})
	if err != nil || name != "report" || output != "new:ok" {
		t.Errorf("new revision used stale compiled template: name=%q output=%q err=%v", name, output, err)
	}
	mu.RLock()
	defer mu.RUnlock()
	if configCalls != 2 || printCalls != 2 {
		t.Errorf("upstream calls: config=%d printcontent=%d, want 2 each", configCalls, printCalls)
	}
}

func TestValidationDoesNotReuseOtherRevision(t *testing.T) {
	service := newService(t,
		func(w http.ResponseWriter, r *http.Request) {
			t.Errorf("unexpected ConfigAppServer request: %s", r.URL.Path)
		},
		func(w http.ResponseWriter, r *http.Request) {
			t.Errorf("unexpected printcontent request: %s", r.URL.Path)
		},
	)
	if err := service.ValidateTemplate(42, "#Value#", "news", "rundown", hash("#Value#"), map[string]any{"Value": "ok"}); err != nil {
		t.Fatalf("first validation: %v", err)
	}
	changed := "#Missing#"
	err := service.ValidateTemplate(42, changed, "news", "rundown", hash(changed), map[string]any{"Value": "ok"})
	var templateErr *diagnostic.TemplateError
	if !errors.As(err, &templateErr) {
		t.Errorf("changed revision should fail on missing field, got %v", err)
	}
	if err := service.ValidateTemplate(42, changed, "news", "rundown", hash(changed), map[string]any{"Missing": "now present"}); err != nil {
		t.Errorf("failed validation poisoned cache: %v", err)
	}
}

func TestPreviewFetchesFreshDataForCachedTemplate(t *testing.T) {
	var mu sync.RWMutex
	previewValue := "first"
	service := newService(t,
		func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/api/news/config/svc/print/template-types/monitoring/preview-data" {
				t.Errorf("unexpected preview path: %s", r.URL.Path)
			}
			mu.RLock()
			defer mu.RUnlock()
			writeJSON(t, w, map[string]any{"Value": previewValue})
		},
		func(w http.ResponseWriter, r *http.Request) {
			t.Errorf("unexpected printcontent request: %s", r.URL.Path)
		},
	)
	output, _, err := service.PreviewTemplate(context.Background(), "#Value#", "news", "monitoring")
	if err != nil || output != "first" {
		t.Fatalf("first preview: output=%q err=%v", output, err)
	}
	mu.Lock()
	previewValue = "second"
	mu.Unlock()
	output, _, err = service.PreviewTemplate(context.Background(), "#Value#", "news", "monitoring")
	if err != nil || output != "second" {
		t.Errorf("cached preview used stale data: output=%q err=%v", output, err)
	}
}

func TestRenderRejectsMismatchedPublishedHashWithoutCaching(t *testing.T) {
	var mu sync.RWMutex
	contentHash := hash("other content")
	service := newService(t,
		func(w http.ResponseWriter, _ *http.Request) {
			mu.RLock()
			defer mu.RUnlock()
			writeJSON(t, w, map[string]any{"name": "report", "template_type": "rundown", "content": "#Value#", "content_hash": contentHash})
		},
		func(w http.ResponseWriter, _ *http.Request) {
			writeJSON(t, w, map[string]any{"Value": "ok"})
		},
	)
	_, _, err := service.RenderTemplate(context.Background(), 42, "news", nil)
	var integrityErr *transformer.TemplateIntegrityError
	if !errors.As(err, &integrityErr) {
		t.Fatalf("expected published content integrity error, got %v", err)
	}
	mu.Lock()
	contentHash = hash("#Value#")
	mu.Unlock()
	_, output, err := service.RenderTemplate(context.Background(), 42, "news", nil)
	if err != nil || output != "ok" {
		t.Errorf("valid published revision after mismatch: output=%q err=%v", output, err)
	}
}

func TestValidationRejectsMismatchedHashAndDoesNotCacheIt(t *testing.T) {
	service := newService(t,
		func(w http.ResponseWriter, r *http.Request) { t.Errorf("unexpected config request: %s", r.URL.Path) },
		func(w http.ResponseWriter, r *http.Request) {
			t.Errorf("unexpected printcontent request: %s", r.URL.Path)
		},
	)
	wrongHash := hash("#Other#")
	err := service.ValidateTemplate(42, "#Value#", "news", "rundown", wrongHash, map[string]any{"Value": "ok"})
	var reqErr *diagnostic.RequestError
	if !errors.As(err, &reqErr) || reqErr.Code != diagnostic.CodeContentHashMismatch {
		t.Fatalf("expected content hash mismatch, got %v", err)
	}
	if err := service.ValidateTemplate(42, "#Other#", "news", "rundown", wrongHash, map[string]any{"Other": "correct"}); err != nil {
		t.Errorf("wrong hash poisoned cache: %v", err)
	}
}

func TestPreviewFailureDoesNotCacheRenderedResult(t *testing.T) {
	var mu sync.RWMutex
	data := map[string]any{"Other": "missing required Value"}
	service := newService(t,
		func(w http.ResponseWriter, _ *http.Request) {
			mu.RLock()
			defer mu.RUnlock()
			writeJSON(t, w, data)
		},
		func(w http.ResponseWriter, r *http.Request) {
			t.Errorf("unexpected printcontent request: %s", r.URL.Path)
		},
	)
	_, _, err := service.PreviewTemplate(context.Background(), "#Value#", "news", "monitoring")
	var templateErr *diagnostic.TemplateError
	if !errors.As(err, &templateErr) {
		t.Fatalf("expected missing field error, got %v", err)
	}
	mu.Lock()
	data = map[string]any{"Value": "restored"}
	mu.Unlock()
	output, _, err := service.PreviewTemplate(context.Background(), "#Value#", "news", "monitoring")
	if err != nil || output != "restored" {
		t.Errorf("preview after bad data: output=%q err=%v", output, err)
	}
}

func TestPreviewRetriesTransientConfigError(t *testing.T) {
	var mu sync.Mutex
	attempts := 0
	service := newService(t,
		func(w http.ResponseWriter, _ *http.Request) {
			mu.Lock()
			defer mu.Unlock()
			attempts++
			if attempts < 3 {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			writeJSON(t, w, map[string]any{"Value": "recovered"})
		},
		func(w http.ResponseWriter, r *http.Request) {
			t.Errorf("unexpected printcontent request: %s", r.URL.Path)
		},
	)
	output, _, err := service.PreviewTemplate(context.Background(), "#Value#", "news", "monitoring")
	if err != nil || output != "recovered" {
		t.Errorf("preview after retry: output=%q err=%v", output, err)
	}
	mu.Lock()
	defer mu.Unlock()
	if attempts != 3 {
		t.Errorf("upstream attempts=%d, want 3", attempts)
	}
}

func TestInvalidCacheLimitsUseDefaults(t *testing.T) {
	newService(t,
		func(w http.ResponseWriter, r *http.Request) { t.Errorf("unexpected config request: %s", r.URL.Path) },
		func(w http.ResponseWriter, r *http.Request) {
			t.Errorf("unexpected printcontent request: %s", r.URL.Path)
		},
	)
	config.PrintCacheMainMaxEntries = -1
	config.PrintCacheMainMaxBytes = -1
	service := transformer.NewTransformService()
	content := "#Value#"
	if err := service.ValidateTemplate(42, content, "news", "monitoring", hash(content), map[string]any{"Value": "ok"}); err != nil {
		t.Errorf("invalid cache configuration should fall back to defaults: %v", err)
	}
}

func TestConcurrentValidationAndRender(t *testing.T) {
	content := "#Value#"
	service := newService(t,
		func(w http.ResponseWriter, _ *http.Request) {
			writeJSON(t, w, map[string]any{"name": "report", "template_type": "rundown", "content": content, "content_hash": hash(content)})
		},
		func(w http.ResponseWriter, _ *http.Request) {
			writeJSON(t, w, map[string]any{"Value": "ok"})
		},
	)
	const workers = 16
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if i%2 == 0 {
				if err := service.ValidateTemplate(42, content, "news", "rundown", hash(content), map[string]any{"Value": "ok"}); err != nil {
					t.Errorf("validation %d: %v", i, err)
				}
				return
			}
			_, output, err := service.RenderTemplate(context.Background(), 42, "news", nil)
			if err != nil || output != "ok" {
				t.Errorf("render %d: output=%q err=%v", i, output, err)
			}
		}(i)
	}
	wg.Wait()
}

func TestRejectsEmptyUpstreamResponses(t *testing.T) {
	for _, tc := range []struct {
		name   string
		target string
		body   string
	}{
		{"published null", "published", "null"},
		{"published empty", "published", ""},
		{"preview null", "preview", "null"},
		{"preview empty", "preview", "   "},
		{"printcontent null", "printcontent", "null"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			content := "literal"
			service := newService(t,
				func(w http.ResponseWriter, r *http.Request) {
					if tc.target == "published" || tc.target == "preview" {
						_, _ = w.Write([]byte(tc.body))
						return
					}
					writeJSON(t, w, map[string]any{"name": "report", "template_type": "rundown", "content": content, "content_hash": hash(content)})
				},
				func(w http.ResponseWriter, _ *http.Request) {
					if tc.target == "printcontent" {
						_, _ = w.Write([]byte(tc.body))
					}
				},
			)
			var err error
			if tc.target == "preview" {
				_, _, err = service.PreviewTemplate(context.Background(), content, "news", "rundown")
			} else {
				_, _, err = service.RenderTemplate(context.Background(), 42, "news", nil)
			}
			if err == nil {
				t.Fatal("empty upstream response was accepted")
			}
		})
	}
}

func TestEmptyPublishedTemplateTypeDoesNotBreakRender(t *testing.T) {
	content := "literal"
	service := newService(t,
		func(w http.ResponseWriter, _ *http.Request) {
			writeJSON(t, w, map[string]any{"name": "report", "content": content, "content_hash": hash(content)})
		},
		func(w http.ResponseWriter, _ *http.Request) { writeJSON(t, w, map[string]any{}) },
	)
	for i := 0; i < 2; i++ {
		name, output, err := service.RenderTemplate(context.Background(), 42, "news", nil)
		if err != nil || name != "report" || output != content {
			t.Fatalf("render %d with an empty template type: name=%q output=%q err=%v", i, name, output, err)
		}
	}
}

func TestRejectsOversizedUpstreamResponse(t *testing.T) {
	const limit = 32 << 20
	service := newService(t,
		func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Length", strconv.Itoa(limit+1))
			_, _ = w.Write([]byte("{}"))
		},
		func(w http.ResponseWriter, r *http.Request) {
			t.Errorf("unexpected printcontent request: %s", r.URL.Path)
		},
	)
	_, _, err := service.PreviewTemplate(context.Background(), "literal", "news", "rundown")
	var fetchErr *transformer.PreviewDataFetchError
	if !errors.As(err, &fetchErr) || !strings.Contains(err.Error(), "лимит") {
		t.Fatalf("expected an upstream response size error, got %v", err)
	}
}

func TestRejectsChunkedOversizedResponse(t *testing.T) {
	const limit = 32 << 20
	service := newService(t,
		func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.(http.Flusher).Flush()
			_, _ = io.CopyN(w, repeatedByteReader{}, limit+1)
		},
		func(w http.ResponseWriter, r *http.Request) {
			t.Errorf("unexpected printcontent request: %s", r.URL.Path)
		},
	)
	_, _, err := service.PreviewTemplate(context.Background(), "literal", "news", "rundown")
	var fetchErr *transformer.PreviewDataFetchError
	if !errors.As(err, &fetchErr) || !strings.Contains(err.Error(), "лимит") {
		t.Fatalf("expected an upstream response size error, got %v", err)
	}
}

func TestArithmeticAtExactJSONIntegerLimit(t *testing.T) {
	service := newService(t,
		func(w http.ResponseWriter, _ *http.Request) {
			_, _ = io.WriteString(w, `{"Value":9007199254740991}`)
		},
		func(w http.ResponseWriter, r *http.Request) {
			t.Errorf("unexpected printcontent request: %s", r.URL.Path)
		},
	)
	output, _, err := service.PreviewTemplate(context.Background(), "<#If Value + 1 == 9007199254740992#>ok<#End#>", "news", "rundown")
	if err != nil || output != "ok" {
		t.Errorf("exact integer arithmetic: output=%q err=%v", output, err)
	}
}

type repeatedByteReader struct{}

func (repeatedByteReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = ' '
	}
	return len(p), nil
}
