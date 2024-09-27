package main

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.azuka.me/connect-playground/internal/app/server"
	"go.azuka.me/connect-playground/internal/pkg/telemetry"
	"go.uber.org/zap"
	"net"
	"net/http"
	"time"
)

func main() {

	logger, err := zap.NewProduction()

	if err != nil {
		panic(err)
	}

	zap.RedirectStdLog(logger)

	defer telemetry.Build("connect-playground-server")()

	h, err := server.Handler(otelzap.New(logger))

	if err != nil {
		panic(err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", 3000))
	if err != nil {
		logger.Fatal("failed to listen", zap.Error(err))
	}

	if err != nil {
		logger.Fatal("unable to initialize server", zap.Error(err))
	}

	s := &http.Server{
		Handler:           h,
		ReadHeaderTimeout: 5 * time.Second,
	}

	logger.Sugar().Infof("server listening at %v", lis.Addr())
	if err = s.Serve(lis); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Fatal("server failed to serve", zap.Error(errors.WithStack(err)))
	}
}
