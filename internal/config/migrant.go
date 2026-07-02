package config

import (
	//	"emobile/internal/config"
	"context"
	"emobile/internal/models"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

func InitMigration(ctx context.Context, cfg Config) (err error) {

	migrant, err := migrate.New(models.MigrationsPath, models.DSN)
	if err != nil {
		pwd, _ := os.Getwd()
		models.Logger.Error("migrate", "MigrationsPath", models.MigrationsPath, "DSN", models.DSN, "ERR", err, "PWD", pwd)
		pureFile, ok := strings.CutPrefix(models.MigrationsPath, "file://")
		if !ok {
			models.Logger.Error("no prefix file://")
		}
		fileInfo, errf := os.Stat(pureFile)
		_ = fileInfo
		if errf != nil {
			models.Logger.Error("no file ", "", pureFile, "err", errf, "pwd", pwd)
		} else {
			models.Logger.Debug("ok", "exist", pureFile)
		}
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer migrant.Close()

	if err := migrant.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}
	models.Logger.Debug("migrate", "", migrant.Log)

	version, dirty, err := migrant.Version()
	if err != nil {
		return fmt.Errorf("failed to get migration version: %w", err)
	}
	models.Logger.Debug("Current migration", "version", version, "dirty", dirty)

	return
}

func getDir(path string) (dirac []string, err error) {
	files, err := os.ReadDir(path)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	for _, file := range files {
		dirac = append(dirac, file.Name())
	}
	return
}

func CheckBase(ctx context.Context, DSN string) (err error) {

	poolConfig, err := pgxpool.ParseConfig(DSN)
	
	if err != nil {
		models.Logger.Error("No", "ParseConfig", err)
		return
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		models.Logger.Error("No", "pgxpool.NewWithConfig", err, "PoolConfig", poolConfig)
		return
	}

	if err = pool.Ping(ctx); err != nil {
		models.Logger.Error("No", "Ping", err)
		return
	}
	pool.Close()

	return
}
