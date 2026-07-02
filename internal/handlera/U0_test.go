package handlera

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"emobile/internal/config"
	"emobile/internal/dbase"
	"emobile/internal/models"

	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TstHand struct {
	suite.Suite
	t   time.Time
	ctx context.Context
	db  *InterStruct
	//	DBEndPoint        string
	postgresContainer testcontainers.Container
}

func (suite *TstHand) SetupSuite() { // выполняется перед тестами

	suite.ctx = context.Background()
	suite.t = time.Now()

	//MigrationsPath = "file://migrations"
	models.MigrationsPath = "file://../../migrations"
	//err := godotenv.Load(models.EnvPath)
	//models.EnvPath = "../../.env"

	// ***************** POSTGREs part begin ************************************
	// Запуск контейнера PostgreSQL
	req := testcontainers.ContainerRequest{
		Image:        "postgres:17",
		Name:         "pcontH", // H - for Handlers
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
	config.Configuration.PageSize = 40

	cfg := config.Configuration
	models.DSN = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)

	err = config.InitMigration(suite.ctx, cfg)
	suite.Require().NoError(err)

	//	suite.DBEndPoint = fmt.Sprintf("postgres://testuser:testpass@%s:%s/testdb", host, port.Port())
	suite.postgresContainer = postgresContainer
	models.Logger.Debug("PostgreSQL", "", host, ":", port.Port())

	db0, err := dbase.NewPostgresPool(context.Background(), models.DSN)

	// suite.db = NewUserHandler(db0)
	suite.db = &InterStruct{Inter: db0}
	suite.Require().NoError(err)

	// ***************** POSTGREs part end ************************************

	models.Logger.Info("SetupTest() --")
}

func (suite *TstHand) TearDownSuite() { // // выполняется после всех тестов
	models.Logger.Info("Spent", "", time.Since(suite.t))
	//suite.dataBase.DB.Close()
	// прикрываем контейнер с БД, для этого и завели переменную в TstHand struct
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

// docker exec -it pcont psql -U testuser -d testdb
