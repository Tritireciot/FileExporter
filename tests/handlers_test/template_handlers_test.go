package handlers

import (
	"fmt"
	"net/http"
	"testing"
)

func TestGetTemplateHandler(t *testing.T) {
	test_data := []struct {
		ID int
		ExpectedStatus int
		ExpectedResponse *map[string]any
	}{
		{
			DBRepoMock.ExpectedId, 
			http.StatusOK,
			&map[string]any{
				"ID": float64(DBRepoMock.ExpectedId),
				"name": DBRepoMock.ExpectedStub,
				"content": DBRepoMock.ExpectedStub,
				"subsystem": DBRepoMock.ExpectedStub,
				"is_active": true,
				"render_data": nil,
				"is_single": false,
			},
		},
		{
			DBRepoMock.UnknownId, 
			http.StatusNotFound,
			nil,
		},
		{
			DBRepoMock.ErrorId, 
			http.StatusInternalServerError,
			nil,
		},
		{
			-1, 
			http.StatusUnprocessableEntity,
			nil,
		},
	}

	for _, set := range test_data {
		recorder := createRequest(
			t,
			MockApp,
			"GET",
			fmt.Sprintf("/db/get_template?id=%d", set.ID), 
			nil,
		)
		checkStatusCode(t, recorder, set.ExpectedStatus)
		if set.ExpectedResponse != nil {
			checkBody(t, recorder, *set.ExpectedResponse)
		}
	}
}


func TestAddTemplateHandler(t *testing.T) {
	test_data := []struct {
		ExpectedStatus int
		ExpectedResponse *map[string]any
		RequestBody *map[string]any
	}{
		{
			http.StatusCreated,
			&map[string]any{
				"ID": float64(DBRepoMock.ExpectedId),
				"name": "test",
				"content": "test",
				"subsystem": "NEWS",
				"is_active": false,
				"render_data": nil,
				"is_single": false,
			},
			&map[string]any{
				"name":  "test",
				"content": "test",
				"subsystem": "NEWS",
			},
		},
		{
			http.StatusInternalServerError,
			nil,
			&map[string]any{
				"name":  DBRepoMock.ErrorName,
				"content": "test",
				"subsystem": "NEWS",
			},
		},
		{
			http.StatusUnprocessableEntity,
			nil,
			&map[string]any{
				"content": "UnprocessableEntity",
			},
		},
	}
	for _, set := range test_data {
		recorder := createRequest(
			t,
			MockApp,
			"POST",
			"/db/add_template", 
			set.RequestBody,
		)
		checkStatusCode(t, recorder, set.ExpectedStatus)
		if set.ExpectedResponse != nil {
			checkBody(t, recorder, *set.ExpectedResponse)
		}
	}
}

func TestDeleteTemplateHandler(t *testing.T) {
	test_data := []struct {
		ID int
		ExpectedStatus int
		ExpectedResponse *map[string]any
	}{
		{
			DBRepoMock.ExpectedId, 
			http.StatusOK,
			&map[string]any{
				"ID": float64(DBRepoMock.ExpectedId),
			},
		},
		{
			DBRepoMock.UnknownId, 
			http.StatusNotFound,
			nil,
		},
		{
			DBRepoMock.ErrorId, 
			http.StatusInternalServerError,
			nil,
		},
		{
			-1, 
			http.StatusUnprocessableEntity,
			nil,
		},
	}

	for _, set := range test_data {
		recorder := createRequest(
			t,
			MockApp,
			"DELETE",
			fmt.Sprintf("/db/delete_template?id=%d", set.ID), 
			nil,
		)
		checkStatusCode(t, recorder, set.ExpectedStatus)
		if set.ExpectedResponse != nil {
			checkBody(t, recorder, *set.ExpectedResponse)
		}
	}
}

func TestGetAllTemplatesHandler(t *testing.T) {
	test_data := []struct {
		ExpectedStatus int
		ExpectedResponse []map[string]any
	}{
		{
			http.StatusOK,
			[]map[string]any{
				{"element_id": float64(1), "element_name": "1", "is_single": false},
				{"element_id": float64(2), "element_name": "2", "is_single": false},
			},
		},
	}

	for _, set := range test_data {
		recorder := createRequest(
			t,
			MockApp,
			"GET",
			"/db/get_all_templates", 
			nil,
		)
		checkStatusCode(t, recorder, set.ExpectedStatus)
		if set.ExpectedResponse != nil {
			checkArrayBody(t, recorder, set.ExpectedResponse)
		}
	}
}