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
    address string
    http    *http.Client
}

func newConfigClient(address string) *configClient {
    return &configClient{
        address: address,
        http:    &http.Client{Timeout: 10 * time.Second},
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

    resp, err := c.http.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return &UpstreamStatusError{StatusCode: resp.StatusCode}
    }

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return fmt.Errorf("чтение ответа ConfigAppServer: %w", err)
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
	requestBody io.Reader,
    result any,
) error {
    delays := [...]time.Duration{100 * time.Millisecond, 300 * time.Millisecond}

    for attempt := 0; attempt <= len(delays); attempt++ {
        if err := ctx.Err(); err != nil {
            return err
        }

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
        var statusErr *UpstreamStatusError
        if errors.As(err, &statusErr) &&
            statusErr.StatusCode == http.StatusNotFound {
            return &TemplateNotFoundError{id: tmpl.ID}
        }
        return fmt.Errorf("получение опубликованного шаблона %d: %w", tmpl.ID, err)
    }
    return nil
}

func (c *configClient) getPrintData(ctx context.Context, context any, tmpl *Template, printData map[string]any) (error) {
	data := map[string]any{
		"context": context,
		"type":    tmpl.Type,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("Получение данных шаблона %s: Не удалось сформировать тело запроса", tmpl.Type)
	}
	endpoint := fmt.Sprintf(
        "http://%s/api/%s/svc/printcontent",
        "mock_backend:8150",//c.address,
        url.PathEscape(tmpl.Subsystem,),
    )
	if err := c.getJSON(ctx, http.MethodPost, endpoint, bytes.NewReader(jsonData), &printData); err != nil {
        var statusErr *UpstreamStatusError
        if errors.As(err, &statusErr) &&
            statusErr.StatusCode == http.StatusNotFound {
            return &TemplateNotFoundError{id: tmpl.ID}
        }
        return fmt.Errorf("Получение данных шаблона %s: %w", tmpl.Type, err)
    }
    return nil
}

func (c *configClient) getPreviewData(
    ctx context.Context,
    subsystem string, 
	templateType string,
	printData map[string]any,
) error {
    endpoint := fmt.Sprintf(
		"http://%s/api/%s/config/svc/print/template-types/%s/preview-data",
		c.address,
		url.PathEscape(subsystem),
		url.PathEscape(templateType),
	)

    if err := c.getJSON(ctx, http.MethodGet, endpoint, nil, &printData); err != nil {
        var statusErr *UpstreamStatusError
        if errors.As(err, &statusErr) &&
            statusErr.StatusCode == http.StatusNotFound {
            return &TemplateTypeNotFoundError{templateType: templateType}
        }
        return fmt.Errorf("получение данных предпросмотра шаблона %s: %w", templateType, err)
    }
    return nil
}