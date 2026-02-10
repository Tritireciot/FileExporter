package main

import (
	"context"
	"fmt"
	"former/internal/config"
	"former/internal/db"
	"former/internal/server"
	"former/internal/transformer"
	"log"
	"os"
)

func main() {
	logger := log.New(os.Stdout, "[SERVICE] ", log.LstdFlags)
	app := server.App{Logger: logger}
	app_config := config.LoadConfig()
	db_repo, err := db.SetupDB(context.Background(), app_config.DB.DB_URL, logger)
	if err != nil {
		logger.Fatal(err.Error())
	}
	transformer_ := transformer.NewTransformService(db_repo)
	app.Init(db_repo, transformer_)
	fmt.Println("App started!!!")
	app.Run(app_config.ConfigureAppUrl())
}