package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	connect "connectrpc.com/connect"
	"connectrpc.com/grpcreflect"
	"github.com/brGuirra/uai/internal/gen/user/v1/userv1connect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func (us *uaiServer) serve() error {
	mux := http.NewServeMux()

	compress1KB := connect.WithCompressMinBytes(1024)

	mux.Handle(userv1connect.NewUserServiceHandler(us, compress1KB))

	mux.Handle(grpcreflect.NewHandlerV1(
		grpcreflect.NewStaticReflector(userv1connect.UserServiceName),
		compress1KB,
	))

	mux.Handle(grpcreflect.NewHandlerV1Alpha(
		grpcreflect.NewStaticReflector(userv1connect.UserServiceName),
		compress1KB,
	))

	srv := &http.Server{
		Addr:              fmt.Sprintf("0.0.0.0:%d", us.config.port),
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

		us.logger.Info(
			"shutting down server",
			"signal", s.String(),
		)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		err := srv.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}

		us.logger.Info(
			"completing background tasks",
			"addr", srv.Addr,
		)

		us.wg.Wait()
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

	us.logger.Info(
		"stopped server",
		"addr", srv.Addr,
	)

	return nil
}
