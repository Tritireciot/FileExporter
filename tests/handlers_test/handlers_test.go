package handlers

import (
	"context"
	"errors"
	"former/internal/db"
	"former/internal/server"
	"former/internal/transformer"
	"log"
	"os"
	"testing"
)

var MockApp server.App
var MockError = errors.New("Mock Error")
var DBRepoMock *MockDB

type MockDB struct {
	ExpectedId int
	ErrorId int
	ExpectedStub string
	UnknownId int
	ErrorName string
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
	} else if tag, ok := element.(*db.Tag); ok {
		tag.Name = repository.ExpectedStub
		tag.Subsystem = repository.ExpectedStub
		tag.Alias = repository.ExpectedStub
	}
	
	return nil
}

func (repository *MockDB) GetAllElements(ctx context.Context, element db.DBModel) (*[]db.ShortElement, error) {
	elements := []db.ShortElement{
		{Element_id: 1, Element_name: "1"},
		{Element_id: 2, Element_name: "2"},
	}
	return &elements, nil
}

func TestMain(m *testing.M) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	MockApp = server.App{Logger: logger}
	DBRepoMock = &MockDB{
		ExpectedId: 1,
		ErrorId: 3,
		ExpectedStub: "ExpectedStub",
		UnknownId: 4,
		ErrorName: "ErrorName",
	}
	transformer_ := transformer.NewTransformService(DBRepoMock)
	MockApp.Init(DBRepoMock, transformer_)
	exit_code := m.Run()
	os.Exit(exit_code)

}
