package main

import (
	"PrintServer/internal/config"
	"PrintServer/internal/db"
	"PrintServer/internal/server"
	"PrintServer/internal/transformer"
	"context"
	"fmt"
	"log"
	"os"
	"time"
)

func main_local() {
	logger := log.New(os.Stdout, "[SERVICE] ", log.LstdFlags)
	app := server.App{Logger: logger}
	app_config := config.LoadConfig()
	var db_repo *db.DBRepository
	for {
		if err := config.GetConfig(); err != nil {
			log.Printf("Failed to get config: %v, retrying...", err)
		} else {
			db_config := config.GetDBConfig()
			if os.Getenv("LOCAL") == "true" {
				db_config.DBname = "backend-db"
				db_config.Host = "db"
				db_config.User = "user"
				db_config.Password = "pass"
				db_config.Schema = "print"
			}
			db_repo, err = db.SetupDB(context.Background(), db_config, logger)
			if err != nil {
				logger.Fatal(err.Error())
				os.Exit(1)
			}
			break
		}
		time.Sleep(time.Second * 5)
	}
	transformer_ := transformer.NewTransformService(db_repo)
	app.Init(db_repo, transformer_)
	fmt.Println("App started!!!")
	app.Run(app_config.ConfigureAppUrl())
}
