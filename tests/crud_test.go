package db

import (
	"context"
	"former/internal/config"
	"former/internal/db"
	"log"
	"os"
	"testing"
)

func createDBRepository() *db.DBRepository {
	logger := log.New(os.Stdout, "[SERVICE] ", log.LstdFlags)
	app_config := config.LoadConfig()
	db_repo, err := db.SetupDB(context.Background(), app_config.DB.DB_URL, logger)
	if err != nil {
		logger.Fatal(err.Error())
		return nil
	}
	return db_repo
}

func TestGetAllElements(t *testing.T) {
	db_repo := createDBRepository();
	template_arr, err := db_repo.GetAllElements(context.Background(), &db.Template{})
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}
	if len(*template_arr) == 0 {
		t.Error("Empty Output")
	}

}