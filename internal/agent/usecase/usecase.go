// Package usecase provides methods for updating all metrics and listing all metrics
package usecase

import (
	"context"

	"github.com/a-palonskaa/metrics-server/internal/models/metrics"
)

// MetricsRepository defines an interface for a metrics storage usecase.
type MetricsRepository interface {
	// Update updates metrics in the repository.
	Update(ctx context.Context, metric metrics.Metrics) error
	// Get returns a metric by its type and name.
	Get(ctx context.Context, mType, name string) (metrics.Metric, error)
	// List returns all stored metrics.
	List(ctx context.Context) []metrics.Metric
	// Close cleans up resources used by the repository.
	Close() error
}
