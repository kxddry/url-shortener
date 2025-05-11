package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/kxddry/url-shortener/internal/config"
	"github.com/kxddry/url-shortener/internal/http-server/handlers/url/redirect"
	"github.com/kxddry/url-shortener/internal/http-server/handlers/url/save"
	mwLogger "github.com/kxddry/url-shortener/internal/http-server/middleware/logger"
	"github.com/kxddry/url-shortener/internal/lib/logger"
	"github.com/kxddry/url-shortener/internal/lib/logger/sl"
	"github.com/kxddry/url-shortener/internal/storage"
	"github.com/kxddry/url-shortener/internal/storage/postgres"
	rds "github.com/kxddry/url-shortener/internal/storage/redis"
	"log/slog"
	"net/http"
	"os"
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
	rdsClient := rds.NewRedisClient(cfg.Redis)
	if err := rdsClient.Connect(); err != nil {
		log.Error("Failed to connect to Redis", sl.Err(err))
		os.Exit(1)
	}
	if err := strg.Connect(); err != nil {
		log.Error("Failed to connect to database", sl.Err(err))
		os.Exit(1)
	}
	if err := strg.New(); err != nil {
		log.Error("Failed to initialize database", sl.Err(err))
		os.Exit(1)
	}
	log.Info("Connected to database", "host", cfg.Storage.Host, "port", cfg.Storage.Port)
	log.Info("Connected to Redis", "host", cfg.Redis.Host, "port", cfg.Redis.Port)

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Post("/url", save.New(log, strg, rdsClient))
	router.Get("/{alias}", redirect.New(log, strg, rdsClient))

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
