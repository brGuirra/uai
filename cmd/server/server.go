package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	connect "connectrpc.com/connect"
	"connectrpc.com/grpcreflect"
	"github.com/brGuirra/uai/internal/config"
	database "github.com/brGuirra/uai/internal/database/sqlc"
	"github.com/brGuirra/uai/internal/gen/user/v1/userv1connect"
	"github.com/brGuirra/uai/internal/smtp"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type appServer struct {
	config *config.Config
	store  database.Querier
	logger *slog.Logger
	mailer *smtp.Mailer
	wg     sync.WaitGroup
	userv1connect.UnimplementedUserServiceHandler
}

func NewAppServer(config *config.Config, store database.Querier, logger *slog.Logger, mailer *smtp.Mailer) appServer {
	return appServer{
		config: config,
		store:  store,
		logger: logger,
		mailer: mailer,
	}
}

func (as *appServer) serve() error {
	mux := http.NewServeMux()

	compress1KB := connect.WithCompressMinBytes(1024)

	mux.Handle(userv1connect.NewUserServiceHandler(as, compress1KB))

	mux.Handle(grpcreflect.NewHandlerV1(
		grpcreflect.NewStaticReflector(userv1connect.UserServiceName),
		compress1KB,
	))

	mux.Handle(grpcreflect.NewHandlerV1Alpha(
		grpcreflect.NewStaticReflector(userv1connect.UserServiceName),
		compress1KB,
	))

	srv := &http.Server{
		Addr:              fmt.Sprintf("0.0.0.0:%d", as.config.Port),
		Handler:           h2c.NewHandler(mux, &http2.Server{}),
		ReadHeaderTimeout: time.Second,
		ReadTimeout:       5 * time.Minute,
		WriteTimeout:      5 * time.Minute,
		MaxHeaderBytes:    8 * 1024, // 8KiB
	}

	shutdownError := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		as.logger.Info(
			"shutting down server",
			"signal", s.String(),
		)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		err := srv.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}

		as.logger.Info(
			"completing background tasks",
			"addr", srv.Addr,
		)

		as.wg.Wait()
		shutdownError <- nil
	}()

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownError
	if err != nil {
		return err
	}

	as.logger.Info(
		"stopped server",
		"addr", srv.Addr,
	)

	return nil
}
