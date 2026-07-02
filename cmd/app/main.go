package main

import (
	"context"
	"emobile/internal/config"
	"emobile/internal/dbase"
	"emobile/internal/handlera"
	"emobile/internal/middlas"
	"emobile/internal/models"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	//	_ "github.com/swaggo/http-swagger"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title Subscription Service API
// @version 1.0
// @description API для управления подписками
// main godoc
// @Summary Запуск приложения
// @Description Основная функция запуска сервиса подписок
// @Produce json
// @Param debug query boolean false "Включить debug-логирование" default(false)
// @Success 200 {string} string "Сервис запущен"
// @Failure 500 {string} string "Ошибка сервера"
func main() {

	ctx := context.Background()

	// уровень логирования по умолчанию Info

	if err := Run(ctx); err != nil {
		models.Logger.Error(err.Error())
	}

}

// Run godoc
// @Summary Запуск сервера API
// @Description Инициализирует конфигурацию и запускает HTTP-сервер с роутингом
// @Accept json
// @Produce json
// @Param ctx query string false "Контекст выполнения"
// @Success 200 {string} string "Сервер запущен"
// @Failure 500 {string} string "Ошибка сервера"
func Run(ctx context.Context) (err error) {

	Level := slog.LevelInfo
	// Если есть флаг -debug
	debugFlag := flag.Bool("debug", false, "установка Минимального уровня логирования DEBUG")
	flag.Parse()
	if *debugFlag {
		Level = slog.LevelDebug
	}
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     Level,
		AddSource: true, // Добавлять информацию об исходном коде
	})
	models.Logger = slog.New(handler)
	slog.SetDefault(models.Logger)
	models.Logger.Debug("Log", "level", Level)

	err = config.InitMigration(ctx, config.Configuration)
	if err != nil {
		//для отладки, спит 900 секунд - время для инспекции контейнера приложения, docker exec -it sapp sh
		models.Logger.Info("sleep ...", "config", config.Configuration)
		time.Sleep(900 * time.Second)
		return
	}

	postgres, err := dbase.NewPostgresPool(context.Background(), models.DSN)
	if err != nil {
		log.Fatalln("NewPostgresPool", "fault", err)
		return
	}
	defer postgres.Close()

	db := &handlera.InterStruct{Inter: postgres}

	router := mux.NewRouter()
	router.HandleFunc("/", db.DBPinger).Methods("GET")
	router.HandleFunc("/add", db.CreateHandler).Methods("POST")
	router.HandleFunc("/read", db.ReadHandler).Methods("POST")
	router.HandleFunc("/list", db.ListHandler).Methods("GET")
	router.HandleFunc("/update", db.UpdateHandler).Methods("PUT")
	router.HandleFunc("/delete", db.DeleteHandler).Methods("DELETE")
	router.HandleFunc("/summa", db.SumHandler).Methods("POST")

	// подключаем middleware логирования
	router.Use(middlas.WithHTTPLogging)
	router.Use(middlas.ErrorLoggerMiddleware)

	router.HandleFunc("/swagger/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		data, _ := os.ReadFile("./docs/swagger.json")
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	})

	// Обработчик для Swagger UI
	router.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL("/swagger/swagger.json"), // Указываем путь к JSON
		httpSwagger.DocExpansion("none"),         // Опционально: схлопывать документацию
	))

	// Контекст для graceful shutdown
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Используем AppHost (или 0.0.0.0) и AppPort для HTTP-сервера
	serverAddr := fmt.Sprintf("%s:%d", config.Configuration.AppHost, config.Configuration.AppPort)

	srv := &http.Server{
		Addr:    serverAddr,
		Handler: router,
	}

	// Запускаем сервер в горутине
	go func() {
		fmt.Printf("\nServer started on %s\n\n", serverAddr)
		models.Logger.Info("Server started", "on", serverAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Ожидаем SIGINT (Ctrl+C) или SIGTERM
	go func() {
		exit := make(chan os.Signal, 1)
		signal.Notify(exit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		<-exit
		cancel() // При получении сигнала отменяем контекст
	}()

	// Блокируемся, пока контекст не отменён
	<-ctx.Done()

	// Graceful shutdown с таймаутом
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		models.Logger.Error("Shutdown", "error", err.Error())
	} else {
		models.Logger.Info("Server stopped gracefully")
	}

	return

}
