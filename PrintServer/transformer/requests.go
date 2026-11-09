package transformer

import (
	"AutoplayX/paths"
	"PrintServer/PrintServer/config"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"
)

type UpstreamStatusError struct {
	StatusCode int
}

type TemplateFetchError struct{ Cause error }

func (e *TemplateFetchError) Error() string {
	return fmt.Sprintf("получение опубликованного шаблона: %v", e.Cause)
}
func (e *TemplateFetchError) Unwrap() error { return e.Cause }

type PreviewDataFetchError struct{ Cause error }

func (e *PreviewDataFetchError) Error() string {
	return fmt.Sprintf("получение данных предпросмотра: %v", e.Cause)
}
func (e *PreviewDataFetchError) Unwrap() error { return e.Cause }

func (e *UpstreamStatusError) Error() string {
	return fmt.Sprintf("ConfigAppServer вернул HTTP %d", e.StatusCode)
}

func retryable(err error) bool {
	var statusErr *UpstreamStatusError
	if errors.As(err, &statusErr) {
		switch statusErr.StatusCode {
		case http.StatusBadGateway,
			http.StatusServiceUnavailable,
			http.StatusGatewayTimeout:
			return true
		}
		return false
	}

	var netErr net.Error
	return errors.As(err, &netErr)
}

type configClient struct {
	address             string
	printContentAddress string
	http                *http.Client
}

const maxUpstreamResponseBytes int64 = 32 << 20

func newConfigClient(address, printContentAddress string) *configClient {
	return &configClient{
		address:             address,
		printContentAddress: printContentAddress,
		http:                &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *configClient) getJSONOnce(
	ctx context.Context,
	method string,
	endpoint string,
	requestBody io.Reader,
	result any,
) error {
	req, err := http.NewRequestWithContext(ctx, method, endpoint, requestBody)
	if err != nil {
		return err
	}
	req.Header.Set(paths.SubAccessTokenTag, config.SubAccessCommKey)
	if method == http.MethodPost {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &UpstreamStatusError{StatusCode: resp.StatusCode}
	}
	if resp.ContentLength > maxUpstreamResponseBytes {
		return fmt.Errorf("ответ внешнего сервиса превышает лимит %d байт", maxUpstreamResponseBytes)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxUpstreamResponseBytes+1))
	if err != nil {
		return fmt.Errorf("чтение ответа ConfigAppServer: %w", err)
	}
	if int64(len(body)) > maxUpstreamResponseBytes {
		return fmt.Errorf("ответ внешнего сервиса превышает лимит %d байт", maxUpstreamResponseBytes)
	}
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return errors.New("внешний сервис вернул пустой JSON-ответ")
	}
	if err := json.Unmarshal(body, result); err != nil {
		return fmt.Errorf("некорректный JSON от ConfigAppServer: %w", err)
	}
	return nil
}

func (c *configClient) getJSON(
	ctx context.Context,
	method string,
	endpoint string,
	bodyBytes []byte,
	result any,
) error {
	delays := [...]time.Duration{100 * time.Millisecond, 300 * time.Millisecond}
	for attempt := 0; attempt <= len(delays); attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		requestBody := bytes.NewReader(bodyBytes)
		err := c.getJSONOnce(ctx, method, endpoint, requestBody, result)
		if err == nil {
			return nil
		}
		if !retryable(err) || attempt == len(delays) {
			return err
		}

		timer := time.NewTimer(delays[attempt])
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
	panic("unreachable")
}

func (c *configClient) getTemplate(
	ctx context.Context,
	tmpl *Template,
) error {
	endpoint := fmt.Sprintf(
		"http://%s/api/%s/config/svc/print/templates/%d/published",
		c.address,
		url.PathEscape(tmpl.Subsystem),
		tmpl.ID,
	)

	if err := c.getJSON(ctx, http.MethodGet, endpoint, nil, tmpl); err != nil {
		return &TemplateFetchError{Cause: err}
	}
	return nil
}

func (c *configClient) getPrintData(ctx context.Context, context any, tmpl *Template) (map[string]any, error) {
	data := map[string]any{
		"context": context,
		"type":    tmpl.Type,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("Получение данных шаблона %s: Не удалось сформировать тело запроса", tmpl.Type)
	}
	endpoint := fmt.Sprintf(
		"http://%s/api/%s/svc/printcontent",
		c.printContentAddress,
		url.PathEscape(tmpl.Subsystem),
	)
	var printData map[string]any
	if err := c.getJSON(ctx, http.MethodPost, endpoint, jsonData, &printData); err != nil {
		var statusErr *UpstreamStatusError
		if errors.As(err, &statusErr) &&
			statusErr.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("Сервис %s недоступен или  шаблон типа %s не существует", tmpl.Subsystem, tmpl.Type)
		}
		return nil, fmt.Errorf("Получение данных шаблона %s: %w", tmpl.Type, err)
	}
	return printData, nil
}

func (c *configClient) getPreviewData(
	ctx context.Context,
	subsystem string,
	templateType string,
) (map[string]any, error) {
	endpoint := fmt.Sprintf(
		"http://%s/api/%s/config/svc/print/template-types/%s/preview-data",
		c.address,
		url.PathEscape(subsystem),
		url.PathEscape(templateType),
	)

	var printData map[string]any

	if err := c.getJSON(ctx, http.MethodGet, endpoint, nil, &printData); err != nil {
		var statusErr *UpstreamStatusError
		if errors.As(err, &statusErr) &&
			statusErr.StatusCode == http.StatusNotFound {
			return nil, &TemplateTypeNotFoundError{templateType: templateType}
		}
		return nil, &PreviewDataFetchError{Cause: err}
	}
	return printData, nil
}
