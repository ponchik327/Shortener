package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	pgxdriver "github.com/wb-go/wbf/dbpg/pgx-driver"
	"github.com/wb-go/wbf/logger"
	wbfredis "github.com/wb-go/wbf/redis"

	rediscache "github.com/ponchik327/Shortener/internal/cache/redis"
	"github.com/ponchik327/Shortener/internal/config"
	"github.com/ponchik327/Shortener/internal/handler"
	"github.com/ponchik327/Shortener/internal/repository"
	"github.com/ponchik327/Shortener/internal/service"
)

const (
	_shutdownTimeout = 15 * time.Second
	_templatesGlob   = "templates/index.html"
	_pgMaxPoolSize   = 10
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	cfgPath := flag.String("config", "config.yaml", "path to config file")
	migrateOnly := flag.Bool("migrate", false, "run database migrations and exit")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	log, err := initLogger(cfg)
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}

	if *migrateOnly {
		return runMigrations(cfg.Database.DSN, log)
	}

	pg, err := pgxdriver.New(cfg.Database.DSN, log, pgxdriver.MaxPoolSize(_pgMaxPoolSize))
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer pg.Close()

	redisClient, err := wbfredis.Connect(wbfredis.Options{
		Address:  cfg.Redis.Address,
		Password: cfg.Redis.Password,
	})
	if err != nil {
		return fmt.Errorf("connect redis: %w", err)
	}

	linkRepo := repository.NewLinkRepository(pg)
	visitRepo := repository.NewVisitRepository(pg)
	linkCache := rediscache.New(redisClient, cfg.Redis.CacheTTL)

	linkSvc := service.NewLinkService(
		linkRepo, linkCache, log,
		cfg.Shortener.CodeLength,
	)
	visitSvc := service.NewVisitService(visitRepo, linkSvc, log)

	h, err := handler.New(linkSvc, visitSvc, _templatesGlob, log)
	if err != nil {
		return fmt.Errorf("init handler: %w", err)
	}

	mux := http.NewServeMux()
	h.Register(mux)

	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      handler.LoggingMiddleware(mux, log),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Info("server starting", "addr", srv.Addr)

	go func() {
		if listenErr := srv.ListenAndServe(); listenErr != nil && !errors.Is(listenErr, http.ErrServerClosed) {
			log.Error("server error", "error", listenErr)
			stop()
		}
	}()

	<-ctx.Done()
	log.Info("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), _shutdownTimeout)
	defer cancel()

	if err = srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}

	log.Info("server stopped")

	return nil
}

func runMigrations(dsn string, log logger.Logger) error {
	m, err := migrate.New("file://migrations", dsn)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}
	defer func() {
		sourceErr, dbErr := m.Close()
		if sourceErr != nil {
			log.Error("migrations: close source", "error", sourceErr)
		}
		if dbErr != nil {
			log.Error("migrations: close db", "error", dbErr)
		}
	}()

	if err = m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Info("migrations: all up to date, nothing to apply")
			return nil
		}

		return fmt.Errorf("apply migrations: %w", err)
	}

	log.Info("migrations: applied successfully")

	return nil
}

func initLogger(cfg *config.Config) (logger.Logger, error) {
	return logger.InitLogger(
		logger.Engine(cfg.Logger.Engine),
		cfg.App.Name,
		cfg.App.Env,
		logger.WithLevel(parseLogLevel(cfg.Logger.Level)),
	)
}

func parseLogLevel(s string) logger.Level {
	switch s {
	case "debug":
		return logger.DebugLevel
	case "warn":
		return logger.WarnLevel
	case "error":
		return logger.ErrorLevel
	default:
		return logger.InfoLevel
	}
}
