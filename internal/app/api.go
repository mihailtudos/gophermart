package app

import (
	"context"
	"errors"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mihailtudos/gophermart/internal/app/accrual"
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

	// Create a context that will be canceled on shutdown signal
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Ensure the cancel function is called

	repos, err := repository.NewRepository(ctx, *cfg.DB)
	if err != nil {
		logger.Log.ErrorContext(context.Background(), "failed to init repository",
			slog.String("err", err.Error()))
		return err
	}

	accrualClient := accrual.New(cfg.Accrual.Address)
	if accrualClient == nil || accrualClient.Address == "" {
		return errors.New("failed to initialize accrual client")
	}

	ss, err := service.NewServices(repos, cfg.Auth, accrualClient)

	// starting the backgorun process
	ss.UpdateOrdersInBackground(ctx, 1*time.Second)

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

	logger.Log.Info("starting server at port: " + cfg.HTTP.Port)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	logger.Log.Info("shutting down server gracefully...")

	// Trigger the context cancellation to stop the background process
	cancel()

	const timeout = 5 * time.Second
	ctx, shutdown := context.WithTimeout(context.Background(), timeout)
	defer shutdown()

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
		// check to ensure only io.Coser records are closed
		if closer, ok := wc.(io.Closer); ok {
			if err := closer.Close(); err != nil {
				log.Printf("failed to close %s file: %v\n", domainKey, err)
			} else {
				log.Printf("successfully closed %s file\n", domainKey)
			}
		} else {
			log.Printf("%s does not implement io.Closer, skipping\n", domainKey)
		}
	}

	return nil
}
