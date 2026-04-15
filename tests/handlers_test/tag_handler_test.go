package handlers

import (
	"fmt"
	"net/http"
	"testing"
)


func TestAddTagHandler(t *testing.T) {
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
				"description": "test",
				"subsystem": "test",
				"alias": "test",
			},
			&map[string]any{
				"name": "test",
				"description": "test",
				"subsystem": "test",
				"alias": "test",
			},
		},
		{
			http.StatusInternalServerError,
			nil,
			&map[string]any{
				"name":  DBRepoMock.ErrorName,
				"description": "test",
				"subsystem": "test",
				"alias": "test",
			},
		},
		{
			http.StatusUnprocessableEntity,
			nil,
			&map[string]any{
				"description": "UnprocessableEntity",
			},
		},
	}
	for _, set := range test_data {
		recorder := createRequest(
			t,
			MockApp,
			"POST",
			"/db/add_tag", 
			set.RequestBody,
		)
		checkStatusCode(t, recorder, set.ExpectedStatus)
		if set.ExpectedResponse != nil {
			checkBody(t, recorder, *set.ExpectedResponse)
		}
	}
}

func TestDeleteTagHandler(t *testing.T) {
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
			fmt.Sprintf("/db/delete_tag?id=%d", set.ID), 
			nil,
		)
		checkStatusCode(t, recorder, set.ExpectedStatus)
		if set.ExpectedResponse != nil {
			checkBody(t, recorder, *set.ExpectedResponse)
		}
	}
}

func TestGetAllTagsHandler(t *testing.T) {
	test_data := []struct {
		ExpectedStatus int
		ExpectedResponse []map[string]any
	}{
		{
			http.StatusOK,
			[]map[string]any{
				{"element_id": float64(1), "element_name": "1"},
				{"element_id": float64(2), "element_name": "2"},
			},
		},
	}

	for _, set := range test_data {
		recorder := createRequest(
			t,
			MockApp,
			"GET",
			"/db/get_all_tags", 
			nil,
		)
		checkStatusCode(t, recorder, set.ExpectedStatus)
		if set.ExpectedResponse != nil {
			checkArrayBody(t, recorder, set.ExpectedResponse)
		}
	}
}