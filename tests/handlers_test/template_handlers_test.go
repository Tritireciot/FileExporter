package handlers

import (
	"fmt"
	"net/http"
	"testing"
)

func TestGetTemplateHandler(t *testing.T) {
	recorder := createRequest(
		t,
		MockApp,
		"GET",
		fmt.Sprintf("/db/get_template?id=%d", DBRepoMock.ExpectedId), 
		nil,
	)
	checkStatusCode(t, recorder, http.StatusOK)

    expected := map[string]any{
		"ID": float64(DBRepoMock.ExpectedId),
		"name": DBRepoMock.ExpectedStub,
		"content": DBRepoMock.ExpectedStub,
	}
	checkBody(t, recorder, expected)
	recorder = createRequest(
		t,
		MockApp,
		"GET",
		fmt.Sprintf("/db/get_template?id=%d", DBRepoMock.UnknownId), 
		nil,
	)
	checkStatusCode(t, recorder,  http.StatusNotFound)
}


func TestAddTemplateHandler(t *testing.T) {
	data := map[string]any{
        "name":  "test",
        "content": "test",
    }
	recorder := createRequest(
		t,
		MockApp,
		"POST",
		"/db/add_template", 
		&data,
	)
	checkStatusCode(t, recorder, http.StatusCreated)

	expected := map[string]any{
		"ID": float64(DBRepoMock.ExpectedId),
		"name": "test",
		"content": "test",
	}
    checkBody(t, recorder, expected)

	data = map[string]any{
        "name":  DBRepoMock.ErrorName,
        "content": "test",
    }
    recorder = createRequest(
		t,
		MockApp,
		"POST",
		"/db/add_template", 
		&data,
	)
	checkStatusCode(t, recorder, http.StatusInternalServerError)
}