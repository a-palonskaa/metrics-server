package repository

import (
	"context"
	"fmt"

	metrics "github.com/a-palonskaa/metrics-server/internal/models/metrics"
)

type MemStorage interface {
	Add(ctx context.Context, mType, name string, val fmt.Stringer)
	Get(ctx context.Context, mType, name string) (fmt.Stringer, bool)
	List(ctx context.Context) []metrics.Metric
	Close() error
}
