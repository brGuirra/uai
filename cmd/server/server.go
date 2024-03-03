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
	"connectrpc.com/validate"
	"github.com/brGuirra/uai/internal/config"
	database "github.com/brGuirra/uai/internal/database/sqlc"
	"github.com/brGuirra/uai/internal/gen/user/v1/userv1connect"
	"github.com/brGuirra/uai/internal/smtp"
	"github.com/brGuirra/uai/internal/token"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type server struct {
	userv1connect.UnimplementedUserServiceHandler
	tokenMaker token.Maker
	store      database.Querier
	config     *config.Config
	logger     *slog.Logger
	mailer     *smtp.Mailer
	wg         sync.WaitGroup
}

func NewServer(config *config.Config, tokenMaker token.Maker, store database.Querier, logger *slog.Logger, mailer *smtp.Mailer) server {
	return server{
		config:     config,
		tokenMaker: tokenMaker,
		store:      store,
		logger:     logger,
		mailer:     mailer,
	}
}

func (s *server) serve() error {
	validateInterceptor, err := validate.NewInterceptor()
	if err != nil {
		return err
	}

	mux := http.NewServeMux()

	compress1KB := connect.WithCompressMinBytes(1024)

	mux.Handle(userv1connect.NewUserServiceHandler(s, compress1KB, connect.WithInterceptors(validateInterceptor)))

	mux.Handle(grpcreflect.NewHandlerV1(
		grpcreflect.NewStaticReflector(userv1connect.UserServiceName),
		compress1KB,
	))

	mux.Handle(grpcreflect.NewHandlerV1Alpha(
		grpcreflect.NewStaticReflector(userv1connect.UserServiceName),
		compress1KB,
	))

	srv := &http.Server{
		Addr:              fmt.Sprintf("0.0.0.0:%d", s.config.Port),
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
		sig := <-quit

		s.logger.Info(
			"shutting down server",
			"signal", sig.String(),
		)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		err := srv.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}

		s.logger.Info(
			"completing background tasks",
			"addr", srv.Addr,
		)

		s.wg.Wait()
		shutdownError <- nil
	}()

	err = srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownError
	if err != nil {
		return err
	}

	s.logger.Info(
		"stopped server",
		"addr", srv.Addr,
	)

	return nil
}
