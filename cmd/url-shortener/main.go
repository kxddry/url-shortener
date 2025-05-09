package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"net/http"
	"os"
	"url-shortener/internal/config"
	"url-shortener/internal/http-server/handlers/url/redirect"
	"url-shortener/internal/http-server/handlers/url/save"
	mwLogger "url-shortener/internal/http-server/middleware/logger"
	"url-shortener/internal/lib/logger"
	"url-shortener/internal/lib/logger/sl"
	"url-shortener/internal/storage"
	"url-shortener/internal/storage/postgres"
)

func main() {
	// init config
	cfg := config.MustLoad()

	// init logger
	log := logger.SetupLogger(cfg.Env)
	log.Info("Starting URL shortener service", "env", cfg.Env)
	log.Info("Host and port", "host", cfg.HTTPServer.Address, "port", cfg.Storage.Port)
	log.Debug("debug messages are enabled")

	// init storage
	var strg storage.Storage
	strg = postgres.NewPostgresStorage(cfg.Storage)
	if err := strg.Connect(); err != nil {
		log.Error("Failed to connect to database", sl.Err(err))
		os.Exit(1)
	}
	if err := strg.New(); err != nil {
		log.Error("Failed to initialize database", sl.Err(err))
		os.Exit(1)
	}
	log.Info("Connected to database", "host", cfg.Storage.Host, "port", cfg.Storage.Port)

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Post("/url", save.New(log, strg))
	router.Get("/{alias}", redirect.New(log, strg))

	log.Info("Starting HTTP server", slog.String("address", cfg.HTTPServer.Address))

	srv := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}
	if err := srv.ListenAndServe(); err != nil {
		log.Error("failed to start server")
	}
	log.Error("server stopped")
}
