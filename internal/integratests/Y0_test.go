package integratests

// Basic imports
import (
	"context"
	"emobile/internal/config"
	"emobile/internal/models"
	"fmt"
	"log"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/suite"
)

type TS struct {
	suite.Suite
	t    time.Time
	ctx  context.Context
	uids [10]string
	host string
}

func (suite *TS) SetupTest() {
	suite.ctx = context.Background()
	suite.t = time.Now()

	dir, err := os.Getwd()
	suite.Require().NoError(err, "pwd", dir)
	models.Logger.Info("Current", "directory:", dir)

	// err = godotenv.Load("./.env")
	suite.Require().NoError(err)
	err = godotenv.Load("../../.env")
	suite.Require().NoError(err, "Setup Test - No .ENV file load")

	cfg := config.Config{
		DBUser:     config.GetEnv("DB_USER", "postgres"),
		DBPassword: config.GetEnv("DB_PASSWORD", ""),
		DBName:     config.GetEnv("DB_NAME", "postgres"),
		DBHost:     "localhost",
		DBPort:     5432,
		AppPort:    8080,
		AppHost:    "localhost",
	}

	models.DSN = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)

	suite.host = fmt.Sprintf("http://%s:%d", cfg.AppHost, cfg.AppPort)

	// PING data base check
	err = config.CheckBase(suite.ctx, models.DSN)
	suite.Require().NoError(err, "No DataBase connection")

	// delete все записи - маска с пустой структурой models.Subscription{}
	httpc := resty.New().SetBaseURL(suite.host)
	req := httpc.R().SetHeader("Content-Type", "application/json").SetDoNotParseResponse(false).
		SetBody(models.Subscription{})
	// раскомментировать если надо обунулять базу перед тестами
	_, err = req.Delete("/delete")
	suite.Require().NoError(err, "DROP")

	for i := range suite.uids {
		suite.uids[i] = uuid.NewString()
	}

}

func (suite *TS) BeforeTest(suiteName, testName string) {
	log.Println("BeforeTest()", suiteName, testName)
}

func (suite *TS) AfterTest(suiteName, testName string) {
	log.Println("AfterTest()", suiteName, testName)
}

func TestExampleTestSuite(t *testing.T) {
	log.Println("before run")

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true, // Добавлять информацию об исходном коде
	})
	models.Logger = slog.New(handler)
	slog.SetDefault(models.Logger)

	suite.Run(t, new(TS))
}
