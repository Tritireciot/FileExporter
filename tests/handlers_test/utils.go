package handlers

import (
	"PrintServer/PrintServer/server"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func createRequest(
	t *testing.T,
	app server.App,
	requestType string,
	url string,
	data *map[string]any,
) *httptest.ResponseRecorder {
	var body io.Reader
	if data == nil {
		body = nil
	} else {
		jsonData, _ := json.Marshal(*data)
		body = bytes.NewBuffer(jsonData)
	}
	req, err := http.NewRequest(requestType, url, body)
	if err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	app.Router.ServeHTTP(recorder, req)
	return recorder
}

func checkStatusCode(
	t *testing.T,
	recorder *httptest.ResponseRecorder,
	expectedStatusCode int,
) {
	if status := recorder.Code; status != expectedStatusCode {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, expectedStatusCode)
	}
}

func checkBody(
	t *testing.T,
	recorder *httptest.ResponseRecorder,
	expectedBody map[string]any,
) {
	var actual map[string]any
	if err := json.NewDecoder(recorder.Body).Decode(&actual); err != nil {
		t.Fatalf("could not decode json: %v", err)
	}
	for key := range actual {
		if actual[key] != expectedBody[key] {
			t.Errorf("handler returned unexpected body param %s : got %v want %v",
				key, actual[key], expectedBody[key])
		}
	}
}

func checkArrayBody(
	t *testing.T,
	recorder *httptest.ResponseRecorder,
	expectedBody []map[string]any,
) {
	var actual []map[string]any
	if err := json.NewDecoder(recorder.Body).Decode(&actual); err != nil {
		t.Fatalf("could not decode json: %v", err)
	}
	for i := range actual {
		for key := range actual[i] {
			if actual[i][key] != expectedBody[i][key] {
				t.Errorf("handler returned unexpected body param %s : got %v want %v",
					key, actual[i][key], expectedBody[i][key])
			}
		}
	}
}
