package dbase

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"emobile/internal/config"
	"emobile/internal/models"

	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TstHand struct {
	suite.Suite
	t        time.Time
	ctx      context.Context
	dataBase *DBstruct
	//	DBEndPoint        string
	postgresContainer testcontainers.Container
}

func (suite *TstHand) SetupSuite() { // выполняется перед тестами

	suite.ctx = context.Background()
	suite.t = time.Now()

	if os.Getenv("DEBUG") == "true" ||
		os.Getenv("DLV_DEBUG") == "1" ||
		os.Getenv("VSCODE_DEBUG") == "true" {
		models.MigrationsPath = "file://migrations"
		models.EnvPath = ".env"
	} else {
		models.MigrationsPath = "file://../../migrations"

		models.EnvPath = "../../.env"
	}

	// ***************** POSTGREs part begin ************************************
	// Запуск контейнера PostgreSQL
	req := testcontainers.ContainerRequest{
		Image:        "postgres:17",
		Name:         "pcontB", // B - для Base
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_PASSWORD": "testpass",
			"POSTGRES_USER":     "testuser",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}

	postgresContainer, err := testcontainers.GenericContainer(suite.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	suite.Require().NoError(err)

	// postgresContainer.Host — возвращает хостнейм или IP-адрес, по которому можно обратиться
	// к запущенному контейнеру с PostgreSQL из тестового приложения.
	// Вместе с MappedPort позволяет сформировать правильный DSN для подключения
	host, err := postgresContainer.Host(suite.ctx)
	suite.Require().NoError(err)

	// контейнеру назначается Внешний (маппированный) порт на хосте, который случайным образом выбирается Docker
	port, err := postgresContainer.MappedPort(suite.ctx, "5432")
	suite.Require().NoError(err)

	config.Configuration.DBPort = port.Int()
	config.Configuration.DBHost = host
	config.Configuration.DBName = "testdb"
	config.Configuration.DBPassword = "testpass"
	config.Configuration.DBUser = "testuser"
	cfg := config.Configuration
	models.DSN = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)

	err = config.InitMigration(suite.ctx, cfg)
	if err != nil {
		return
	}

	//	suite.DBEndPoint = fmt.Sprintf("postgres://testuser:testpass@%s:%s/testdb", host, port.Port())
	suite.postgresContainer = postgresContainer
	models.Logger.Debug("PostgreSQL", "", host, ":", port.Port())

	suite.dataBase, err = NewPostgresPool(suite.ctx, models.DSN)
	suite.Require().NoError(err)

	// ***************** POSTGREs part end ************************************

	models.Logger.Info("SetupTest() --")
}

func (suite *TstHand) TearDownSuite() { // // выполняется после всех тестов
	models.Logger.Info("Spent", "", time.Since(suite.t))
	// сначала закрываем БД
	suite.dataBase.DB.Close()
	// прикрываем контейнер с БД, только для этого и завели переменную в TstHand struct
	suite.postgresContainer.Terminate(suite.ctx)
}

func TestHandlersSuite(t *testing.T) {
	testBase := new(TstHand)
	testBase.ctx = context.Background()

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true, // Добавлять информацию об исходном коде
	})
	models.Logger = slog.New(handler)
	slog.SetDefault(models.Logger)

	suite.Run(t, testBase)

}

// MakeTT используется в функциях тестов, преобразует поля со строковыми датами в time.Time
func MakeTT(sub *models.Subscription) (err error) {

	sub.Sdt, err = models.ParseDate(sub.Start_date)
	if err != nil {
		return
	}
	sub.Edt, err = models.ParseDate(sub.End_date)
	return
}

// docker exec -it pcont psql -U testuser -d testdb
