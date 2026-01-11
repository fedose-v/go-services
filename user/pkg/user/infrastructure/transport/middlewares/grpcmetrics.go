package middlewares

import (
	"context"
	"errors"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"user/pkg/user/domain/model"
)

func NewGRPCMetricsMiddleware() grpc.UnaryServerInterceptor {
	vec := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "service_requests_total",
			Help: "Application total request",
		},
		[]string{"method", "code"},
	)
	prometheus.MustRegister(vec)

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		start := time.Now()

		resp, err = handler(ctx, req)

		var code codes.Code
		switch {
		case err == nil:
			code = codes.OK
		case errors.Is(err, model.ErrUserNotFound):
			code = codes.NotFound
		case err != nil:
			code = codes.Internal
		}

		duration := time.Since(start).Seconds()

		vec.
			WithLabelValues(info.FullMethod, code.String()).
			Observe(duration)

		return resp, err
	}
}
