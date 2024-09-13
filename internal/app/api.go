package app

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mihailtudos/gophermart/internal/config"
	"github.com/mihailtudos/gophermart/internal/delivery"
	"github.com/mihailtudos/gophermart/internal/logger"
	"github.com/mihailtudos/gophermart/internal/repository"
	"github.com/mihailtudos/gophermart/internal/server"
	"github.com/mihailtudos/gophermart/internal/service"
)

func Run(configPath string) error {
	cfg, err := config.NewConfig(configPath)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	repos, err := repository.NewRepository(ctx, *cfg.DB)
	if err != nil {
		logger.Log.ErrorContext(context.Background(), "failed to init repository",
			slog.String("err", err.Error()))
		return err
	}

	ss, err := service.NewServices(repos, cfg.Auth)

	if err != nil {
		logger.Log.ErrorContext(context.Background(), "failed to init services",
			slog.String("err", err.Error()))
		return err
	}

	handler := delivery.NewHandler(ss)
	srv := server.NewServer(cfg, handler.Init())

	go func() {
		if err := srv.Run(); !errors.Is(err, http.ErrServerClosed) {
			logger.Log.Error("error occurred while running http server: %s\n", slog.String("err", err.Error()))
		}
	}()

	logger.Log.Info("starting server at port: " + cfg.Http.Port)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	const timeout = 5 * time.Second
	ctx, shutdown := context.WithTimeout(context.Background(), timeout)
	defer shutdown()

	logger.Log.Info("shutting down server gracefully...")

	// stopping server
	if err := srv.Stop(ctx); err != nil {
		logger.Log.Error("failed to stop server: %v", slog.String("err", err.Error()))
	}

	// closing storage connection
	if err := repos.Close(); err != nil {
		logger.Log.Error("failed to close db connection", slog.String("err", err.Error()))
	}

	logger.Log.Info("Server exiting, closing the log files")

	for domainKey, wc := range cfg.ToClose {
		if err := wc.Close(); err != nil {
			log.Printf("failed to close %s file\n", domainKey)
		}
		log.Printf("successfully closed %s file\n", domainKey)
	}

	return nil
}
