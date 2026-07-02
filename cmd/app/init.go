package main

import (
	"emobile/internal/config"
	"emobile/internal/models"
	"fmt"
	"log"
	"os"
)

func init() {

	var err error

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	//
	// если сервер запущен в контейнере, в нём есть переменная окружения MIGRATIONS_PATH
	enva, exists := os.LookupEnv("MIGRATIONS_PATH")
	if exists {
		// задаём путь к файлам миграции в самом контейнере
		models.MigrationsPath = enva
	} else {
		// если нет "MIGRATIONS_PATH" значит приложение запущено не в контейнере и хост localhost
		cfg.DBHost = "localhost"
	}

	models.DSN = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)

	config.Configuration = cfg

	log.Println("DB ok", "DSN", models.DSN)

	}
