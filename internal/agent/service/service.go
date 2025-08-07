// Package service provides Handler interface for metrics-sender agent
package service

import (
	"context"
)

type Handler interface {
	SendMetrics(ctx context.Context) error
	UpdateRuntimeMetrics(ctx context.Context)
	UpdateSystemMetrics(ctx context.Context)
}
