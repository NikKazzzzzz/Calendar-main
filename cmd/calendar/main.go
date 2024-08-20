package main

import (
	"github.com/NikKazzzzzz/Calendar-main/internal/http-server/handlers"
	"github.com/NikKazzzzzz/Calendar-main/internal/lib/logger/sl"
	"github.com/NikKazzzzzz/Calendar-main/internal/storage/mongodb"
	"github.com/NikKazzzzzz/Calendar-main/monitoring"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/NikKazzzzzz/Calendar-main/internal/config"
	mwLogger "github.com/NikKazzzzzz/Calendar-main/internal/http-server/middleware/logger"
	mwPrometheus "github.com/NikKazzzzzz/Calendar-main/internal/http-server/middleware/prometheus"
	"github.com/NikKazzzzzz/Calendar-main/internal/lib/logger/handlers/slogpretty"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info("starting calendar",
		slog.String("env", cfg.Env),
		slog.String("databaseName", cfg.DatabaseName),
	)
	log.Debug("debug message are enabled")

	// Проверяем, заданы ли username и password в конфигурации. Если нет, используем значения из переменных окружения.
	username := cfg.Username
	if username == "" {
		username = os.Getenv("MONGO_USERNAME")
	}

	password := cfg.Password
	if password == "" {
		password = os.Getenv("MONGO_PASSWORD")
	}

	// Замена username и password в строке подключения
	mongoDSN := strings.Replace(cfg.MongoDSN, "username", username, 1)
	mongoDSN = strings.Replace(mongoDSN, "password", password, 1)

	log.Debug("Connecting to MongoDB using DSN:", slog.String("mongo_dsn", mongoDSN))

	storage, err := mongodb.New(mongoDSN, cfg.DatabaseName, username, password, log)
	if err != nil {
		log.Error("failed to init MongoDB storage", sl.Err(err))
		os.Exit(1)
	}

	monitoring.Init()

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(mwLogger.New(log))
	router.Use(mwPrometheus.PrometheusMiddleware)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Handle("/metrics", promhttp.Handler())

	eventHandler := handlers.NewEventHandler(storage, log)

	router.Post("/events", eventHandler.CreateEvent)
	router.Put("/events/{id}", eventHandler.UpdateEvent)
	router.Delete("/events/{id}", eventHandler.DeleteEvent)
	router.Get("/events/{id}", eventHandler.GetEventByID)
	router.Get("/events", eventHandler.GetAllEvents)

	log.Info("starting server", slog.String("address", cfg.Address))

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Error("failed to start server", sl.Err(err))
	}

}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	default:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)

	}

	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
