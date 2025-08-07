package workerpool_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	workerpool "github.com/a-palonskaa/metrics-server/pkg/worker_pool"
)

func TestWorkerPool_TaskCancelledByContext(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	wp := workerpool.New(1, ctx)
	defer wp.Close()

	wp.AddTask(func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second): // never reached
			return nil
		}
	})

	select {
	case err := <-wp.Result():
		require.ErrorIs(t, err, context.DeadlineExceeded)
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timeout waiting for result")
	}
}

func TestWorkerPool(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	wp := workerpool.New(1, ctx)
	defer wp.Close()

	wp.AddTask(func(ctx context.Context) error {
		return nil
	})

	select {
	case err := <-wp.Result():
		require.NoError(t, err)
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timeout waiting for result")
	}
}
