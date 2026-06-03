package handlers

import (
	"PrintServer/internal/db"
	"PrintServer/internal/server"
	"context"
	"errors"
)

var MockApp server.App
var MockError = errors.New("Mock Error")
var DBRepoMock *MockDB
var TransformerServiceMock *MockTransformer

type MockDB struct {
	ExpectedId   int
	ErrorId      int
	ExpectedStub string
	UnknownId    int
	ErrorName    string
}

func (repository *MockDB) AddTag(ctx context.Context, tag *db.Tag) error {
	if tag.Name == repository.ErrorName {
		return MockError
	}
	tag.ID = repository.ExpectedId
	return nil
}

func (repository *MockDB) AddTemplate(ctx context.Context, template *db.Template) error {
	if template.Name == repository.ErrorName {
		return MockError
	}
	template.ID = repository.ExpectedId
	return nil
}

func (repository *MockDB) DeleteElement(ctx context.Context, element db.DBModel, column string) error {
	if element.GetID() == repository.UnknownId {
		return db.ErrElementNotFound
	} else if element.GetID() == repository.ErrorId {
		return MockError
	}
	return nil
}

func (repository *MockDB) GetElement(ctx context.Context, element db.DBModel, column string) error {
	if element.GetID() == repository.UnknownId {
		return db.ErrElementNotFound
	} else if element.GetID() == repository.ErrorId {
		return MockError
	}
	if template, ok := element.(*db.Template); ok {
		template.Name = repository.ExpectedStub
		template.Content = repository.ExpectedStub
		template.Subsystem = repository.ExpectedStub
		template.IsActive = true
		template.RenderData = nil
		template.IsSingle = false
	} else if tag, ok := element.(*db.Tag); ok {
		tag.Name = repository.ExpectedStub
		tag.Subsystem = repository.ExpectedStub
		tag.Alias = repository.ExpectedStub
		tag.IsActive = true
	}

	return nil
}

func (repository *MockDB) GetAllElements(ctx context.Context, element db.DBModel, subsystem string, isActive bool) (*[]db.ShortElement, error) {
	elements := []db.ShortElement{
		{Element_id: 1, Element_name: "1"},
		{Element_id: 2, Element_name: "2"},
	}
	return &elements, nil
}

type MockTransformer struct {
	ExpectedId int
	ErrorId    int
}

func (service *MockTransformer) RenderTemplate(ctx context.Context, template_id int, raw_data []byte) (string, error) {
	if template_id != service.ErrorId {
		return "RenderedTemplate", nil
	}
	return "", MockError
}
