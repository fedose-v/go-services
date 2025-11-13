package main

import (
	"context"
	"net"
	"time"

	applogger "gitea.xscloud.ru/xscloud/golib/pkg/application/logging"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"

	"user/api/server/userinternal"
	"user/pkg/infrastructure/transport"
)

const shutdownTimeout = 30 * time.Second

func service(
	config *config,
	logger applogger.Logger,
	closer *multiCloser,
) *cli.Command {
	return &cli.Command{
		Name:  "service",
		Usage: "Runs the gRPC service",
		Action: func(c *cli.Context) error {
			connContainer, err := newConnectionsContainer(config, logger, closer)
			if err != nil {
				return errors.Wrap(err, "failed to init connections")
			}

			container, err := newDependencyContainer(config, connContainer)
			if err != nil {
				return errors.Wrap(err, "failed to init dependencies")
			}
			return startGRPCServer(c.Context, config, logger, container)
		},
	}
}

func startGRPCServer(
	ctx context.Context,
	config *config,
	logger applogger.Logger,
	_ *dependencyContainer,
) error {
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(makeGrpcUnaryInterceptor(logger)))

	// TODO: зарегистрировать свой сервер вместо шаблонного
	userinternal.RegisterMicroserviceTemplateInternalServiceServer(grpcServer, transport.NewUserAPI())

	listener, err := net.Listen("tcp", config.ServeGRPCAddress)
	if err != nil {
		return errors.Wrapf(err, "failed to listen on %s", config.ServeGRPCAddress)
	}
	logger.Info("gRPC server listening on", config.ServeGRPCAddress)

	errCh := make(chan error, 1)
	go func() {
		errCh <- grpcServer.Serve(listener)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		logger.Info("Shutdown signal received, stopping gRPC server...")
		shutdownGRPCServer(grpcServer, logger)
		return nil
	}
}

func shutdownGRPCServer(server *grpc.Server, logger applogger.Logger) {
	done := make(chan struct{})
	go func() {
		server.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		logger.Info("gRPC server stopped gracefully")
	case <-time.After(shutdownTimeout):
		logger.Error(errors.Errorf("graceful shutdown timed out after %v, forcing stop", shutdownTimeout))
		server.Stop()
	}
}

func makeGrpcUnaryInterceptor(logger applogger.Logger) grpc.UnaryServerInterceptor {
	loggerInterceptor := transport.MakeLoggerServerInterceptor(logger)
	errorInterceptor := transport.ErrorInterceptor{Logger: logger}
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		resp, err = loggerInterceptor(ctx, req, info, handler)
		return resp, errorInterceptor.TranslateGRPCError(err)
	}
}
