package config

import (
	"emobile/internal/models"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DBUser     string
	DBPassword string
	DBName     string
	DBHost     string
	DBPort     int
	AppPort    int
	AppHost    string
	PageSize   int
}

var Configuration Config

func Load() (Config, error) {
	// Загружаем .env файл
	// Load will read your env file(s) and load them into ENV for this process.
	err := godotenv.Load(models.EnvPath)
	// err := godotenv.Load("../../.env")
	if err != nil {
		dir, errd := os.Getwd()
		log.Println("Warning: couldn't load .env ", "Error:", err, "pwd", dir, "err os.Getwd", errd)
		// It's important to note that it WILL NOT OVERRIDE an env variable
		// that already exists - consider the .env file to set dev vars or sensible defaults.
		// Не прерываем выполнение, так как переменные могут быть установлены в окружении
		// но и они на самом деле установлены - .env прочитан docker compose
		//
	}

	// Парсим порт приложения
	appPort, err := strconv.Atoi(GetEnv("APP_PORT", "8080"))
	if err != nil {
		return Config{}, err
	}

	// Парсим порт БД
	dbPort, err := strconv.Atoi(GetEnv("DB_PORT", "5432"))
	if err != nil {
		return Config{}, err
	}

	pageSizeStr := GetEnv("PAGE_SIZE", "20")
	pagesz, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		return Config{}, err
	}
	cfg := Config{
		DBUser:     GetEnv("DB_USER", "postgres"),
		DBPassword: GetEnv("DB_PASSWORD", ""),
		DBName:     GetEnv("DB_NAME", "postgres"),
		DBHost:     GetEnv("DB_HOST", "localhost"),
		DBPort:     dbPort,
		AppPort:    appPort,
		AppHost:    GetEnv("APP_HOST", "0.0.0.0"),
		PageSize:   pagesz,
	}

	// models.Logger.Info("Environment ", "vars", cfg)

	return cfg, nil
}

// GetEnv возвращает значение переменной окружения или значение по умолчанию
func GetEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
