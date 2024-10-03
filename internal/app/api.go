package app

import (
	"context"
	"errors"
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
	"github.com/mihailtudos/gophermart/internal/service/auth"
)

const (
	shutdownTimeout     = 5 * time.Second
	updateOrdersTimeout = 1 * time.Minute
)

func Run() error {
	cfg := config.NewConfig()
	logger.New(nil, cfg.Logger.Level)

	// Create a context that will be canceled on shutdown signal
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	repos, err := repository.NewRepository(ctx, cfg.DB)
	if err != nil {
		logger.Log.ErrorContext(ctx,
			"failed to init repository",
			slog.String("err", err.Error()))
		return err
	}

	accrualClient := accrual.New(cfg.Accrual.Address)
	if accrualClient == nil || accrualClient.Address == "" {
		return errors.New("failed to initialize accrual client")
	}

	tms, err := auth.NewManager(cfg.Auth.JWT)
	if err != nil {
		return err
	}

	userService, err := service.NewUserService(repos.UserRepo, tms)
	if err != nil {
		return err
	}

	ss, err := service.NewServices(userService, tms, accrualClient)

	// starting the backgorun process
	ss.UpdateOrdersInBackground(ctx, updateOrdersTimeout)

	if err != nil {
		logger.Log.ErrorContext(ctx,
			"failed to init services",
			slog.String("err", err.Error()))
		return err
	}

	srv := server.NewServer(cfg.HTTP, delivery.NewHandler(ss.UserService, ss.UserService))

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

	ctx, shutdown := context.WithTimeout(ctx, shutdownTimeout)
	defer shutdown()

	// stopping server
	if err := srv.Stop(ctx); err != nil {
		logger.Log.Error("failed to stop server: %v", slog.String("err", err.Error()))
	}

	// closing storage connection
	if err := repos.Close(); err != nil {
		logger.Log.Error("failed to close db connection", slog.String("err", err.Error()))
	}

	return nil
}
