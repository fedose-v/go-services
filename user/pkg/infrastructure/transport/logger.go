package transport

import (
	"context"
	"time"

	applogger "gitea.xscloud.ru/xscloud/golib/pkg/application/logging"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type ErrorInterceptor struct {
	Logger applogger.Logger
}

func (i ErrorInterceptor) TranslateGRPCError(err error) error {
	if err == nil {
		return nil
	}
	// if already a GRPC error return unchanged
	if _, ok := status.FromError(err); ok {
		return err
	}

	return status.Error(getGRPCCode(err), err.Error())
}

func MakeLoggerServerInterceptor(logger applogger.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		start := time.Now()

		resp, err = handler(ctx, req)

		duration := time.Since(start).String()
		fields := log.Fields{
			"duration": duration,
			"route":    info.FullMethod,
		}

		loggerWithFields := logger.WithFields(applogger.Fields(fields))
		if err == nil {
			loggerWithFields.Info("call finished")
		} else {
			loggerWithFields.Warning(err, "call failed")
		}
		return resp, err
	}
}
