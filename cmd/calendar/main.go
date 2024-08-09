package main

import (
	"github.com/NikKazzzzzz/Calendar-main/internal/http-server/handlers"
	"github.com/NikKazzzzzz/Calendar-main/internal/lib/logger/sl"
	"github.com/NikKazzzzzz/Calendar-main/internal/storage/postgres"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/NikKazzzzzz/Calendar-main/internal/config"
	mwLogger "github.com/NikKazzzzzz/Calendar-main/internal/http-server/middleware/logger"
	"github.com/NikKazzzzzz/Calendar-main/internal/lib/logger/handlers/slogpretty"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
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
	)
	log.Debug("debug message are enabled")

	log.Debug("Using DSN:", slog.String("dsn", cfg.StorageDSN))

	storage, err := postgres.New(cfg.StorageDSN, log)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	eventHandler := handlers.NewEventHandler(storage, log)

	router.Post("/events", eventHandler.CreateEvent)
	router.Put("/events/{id}", eventHandler.UpdateEvent)
	router.Delete("/events/{id}", eventHandler.DeleteEvent)
	router.Get("/events/{id}", eventHandler.GetEventByID)

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
