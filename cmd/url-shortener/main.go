package main

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	ssogrpc "github.com/kxddry/url-shortener/internal/clients/sso/grpc"
	"github.com/kxddry/url-shortener/internal/config"
	del "github.com/kxddry/url-shortener/internal/http-server/handlers/url/delete"
	"github.com/kxddry/url-shortener/internal/http-server/handlers/url/homepage"
	"github.com/kxddry/url-shortener/internal/http-server/handlers/url/login"
	"github.com/kxddry/url-shortener/internal/http-server/handlers/url/redirect"
	"github.com/kxddry/url-shortener/internal/http-server/handlers/url/register"
	"github.com/kxddry/url-shortener/internal/http-server/handlers/url/save"
	mwLogger "github.com/kxddry/url-shortener/internal/http-server/middleware/logger"
	"github.com/kxddry/url-shortener/internal/lib/logger"
	"github.com/kxddry/url-shortener/internal/lib/logger/sl"
	"github.com/kxddry/url-shortener/internal/storage/postgres"
	rds "github.com/kxddry/url-shortener/internal/storage/redis"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// init config
	cfg := config.MustLoad()

	// init logger
	log := logger.SetupLogger(cfg.Env)
	log.Info("Starting URL shortener service", "env", cfg.Env)
	log.Info("Host and port", "host", cfg.HTTPServer.Address, "port", cfg.Storage.Port)
	log.Debug("debug messages are enabled")

	// init SSO client
	sso := cfg.Clients.SSO
	ssoClient, err := ssogrpc.New(context.Background(), log, sso.Address, sso.Timeout, sso.Retries)
	if err != nil {
		log.Error("Error initializing SSO client", sl.Err(err))
		os.Exit(1)
	}

	cfg.App.ID, err = ssoClient.AppID(context.Background(), cfg.App.Name, cfg.App.Secret)
	if err != nil {
		log.Error("failed to get appId", sl.Err(err))
		os.Exit(1)
	}

	ctx := context.Background()

	// init storage
	store, err := postgres.New(cfg.Storage)
	if err != nil {
		log.Error("Failed to connect to database", sl.Err(err))
		os.Exit(1)
	}
	redis, err := rds.New(cfg.Redis)
	if err != nil {
		log.Error("Failed to connect to Redis", sl.Err(err))
		os.Exit(1)
	}
	log.Info("Connected to database", "host", cfg.Storage.Host, "port", cfg.Storage.Port)
	log.Info("Connected to Redis", "host", cfg.Redis.Host, "port", cfg.Redis.Port)

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Post("/url", save.New(log, store, redis, cfg))
	router.Get("/", homepage.Url(log, cfg))
	router.Get("/url", homepage.Url(log, cfg))

	router.Get("/login", homepage.Login(cfg))
	router.Post("/login", login.New(ctx, log, cfg, ssoClient))

	router.Get("/register", homepage.Register(cfg))
	router.Post("/register", register.New(ctx, log, ssoClient))

	router.Get("/{alias}", redirect.New(log, store, redis))
	router.Delete("/{alias}", del.New(ctx, log, cfg, store, redis, ssoClient))

	log.Info("Starting HTTP server", slog.String("address", cfg.HTTPServer.Address))

	srv := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}
	go func() {
		err = srv.ListenAndServe()
		if err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return
			}
			log.Error("Failed to start HTTP server", sl.Err(err))
			os.Exit(1)
		}
	}()

	// graceful shutdown

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Info("Shutting down HTTP server")
	_ = srv.Shutdown(context.Background())
	log.Info("Shutting down Redis server")
	_ = redis.Close()
	log.Info("Shutting down SQL connection")
	_ = store.Close()
	log.Info("Application stopped")

}
