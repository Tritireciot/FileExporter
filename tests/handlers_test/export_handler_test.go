package handlers

import (
	"net/http"
	"testing"
)

func TestExportTemplateHandler(t *testing.T) {
	test_data := []struct {
		ExpectedStatus   int
		ExpectedResponse *map[string]any
		RequestBody      *map[string]any
	}{
		{
			http.StatusOK,
			&map[string]any{
				"ID":          float64(TransformerServiceMock.ExpectedId),
				"name":        "Title",
				"content":     "RenderedTemplate",
				"subsystem":   "news",
				"is_active":   false,
				"render_data": nil,
				"is_single":   false,
			},
			&map[string]any{
				"template_id": TransformerServiceMock.ExpectedId,
				"data": map[string]any{
					"command": "update",
				},
				"subsystem": "news",
			},
		},
		{
			http.StatusInternalServerError,
			nil,
			&map[string]any{
				"template_id": TransformerServiceMock.ErrorId,
				"data": map[string]any{
					"command": "update",
				},
				"subsystem": "news",
			},
		},
		{
			http.StatusUnprocessableEntity,
			nil,
			&map[string]any{
				"template_id": 0,
				"data":        map[string]any{},
				"subsystem":   "",
			},
		},
	}
	for _, set := range test_data {
		recorder := createRequest(
			t,
			MockApp,
			"POST",
			"/api/core/print/export",
			set.RequestBody,
		)
		checkStatusCode(t, recorder, set.ExpectedStatus)
		if set.ExpectedResponse != nil {
			checkBody(t, recorder, *set.ExpectedResponse)
		}
	}
}
