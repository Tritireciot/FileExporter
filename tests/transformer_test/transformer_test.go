package transformer

import (
	"context"
	"former/internal/config"
	"former/internal/db"
	"former/internal/transformer"
	"log"
	"os"
	"testing"
)

var test_db_repo *db.DBRepository
var test_reshaper *transformer.Reshaper
var test_context context.Context

func createDBRepository() *db.DBRepository {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	app_config := config.LoadConfig()
	db_repo, err := db.SetupDB(context.Background(), app_config.DB.DB_URL, logger)
	if err != nil {
		logger.Fatal(err.Error())
		return nil
	}
	return db_repo
}

func TestMain(m *testing.M) {
	test_db_repo = createDBRepository()
	test_context = context.Background()
	defer test_db_repo.TearDown()
	test_reshaper = transformer.NewReshaper(test_db_repo)
	exit_code := m.Run()
	os.Exit(exit_code)

}
