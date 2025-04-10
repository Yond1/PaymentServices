package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"paymentSystem/internal/config"
	"paymentSystem/internal/database"
	"paymentSystem/internal/repository"
	"paymentSystem/internal/router"
	"syscall"
	"time"
)

func main() {

	cfg := config.GetConfig()

	log := initLogger(cfg)

	log.Info("connecting to db")

	storage, err := database.NewStorage(cfg)
	if err != nil {
		log.Error("error connecting to db", "error", err)
		os.Exit(1)
	}

	storage.Postgres.SetMaxOpenConns(50)
	storage.Postgres.SetMaxIdleConns(25)
	storage.Postgres.SetConnMaxLifetime(5 * time.Minute)

	err = runMigrations(storage)
	if err != nil {
		log.Error("error running migrations", "error", err)
		os.Exit(1)
	}

	defer func() {
		err := storage.Postgres.Close()
		if err != nil {
			log.Error("error closing db", "error", err)
			os.Exit(1)
		}
	}()

	log.Info("create repository")

	repo, err := repository.NewRepository(storage)
	if err != nil {
		log.Error("error creating repository", "error", err)
		os.Exit(1)
	}

	log.Info("create router")

	g := gin.Default()

	g.Use(gin.Recovery())

	router.SetupRoutes(g, repo)

	address := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)

	log.Info("register routes")

	server := &http.Server{Addr: address, Handler: g}

	go func() {

		log.Info("Starting server", "address", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Server error", "error", err)
			os.Exit(1)
		}

	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("Server shutdown error", "error", err)
	}

	log.Info("Server stopped")

}

func initLogger(cfg *config.Config) *slog.Logger {
	var log *slog.Logger
	switch cfg.LevelLog {
	case "dev":
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case "stage":
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	case "prod":
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	default:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	return log
}

func runMigrations(db *database.Storage) error {
	_, err := db.Postgres.Exec("CREATE TABLE IF NOT EXISTS wallets(wallet_id UUID PRIMARY KEY, balance INTEGER)")
	if err != nil {
		return err
	}
	db.Postgres.Exec("INSERT INTO wallets(wallet_id, balance) VALUES ('00000000-0000-0000-0000-000000000000', 5000)")
	return nil
}
