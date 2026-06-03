package handlers

import (
	"PrintServer/internal/server"
	"log"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	MockApp = server.App{Logger: logger}
	DBRepoMock = &MockDB{
		ExpectedId:   1,
		ErrorId:      3,
		ExpectedStub: "ExpectedStub",
		UnknownId:    4,
		ErrorName:    "ErrorName",
	}
	TransformerServiceMock = &MockTransformer{
		ExpectedId: 1,
		ErrorId:    3,
	}
	MockApp.Init(DBRepoMock, TransformerServiceMock)
	exit_code := m.Run()
	os.Exit(exit_code)

}
