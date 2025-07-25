package usecase

import (
	"context"

	"github.com/a-palonskaa/metrics-server/internal/models/metrics"
)

type Connector interface {
	Ping(ctx context.Context) error
}

type MetricsRepository interface {
	Update(ctx context.Context, metric metrics.Metrics) error
	Get(ctx context.Context, mType, name string) (metrics.Metric, error)
	List(ctx context.Context) []metrics.Metric
	Close() error
}
