package usecase

import (
	"context"
	"errors"

	"github.com/rs/zerolog/log"
)

var (
	// ErrConnNotInitialized is returned when a Ping operation is attempted without an initialized Connector.
	ErrConnNotInitialized = errors.New("connector does not exist")
)

type Ping struct {
	conn Connector
}

// NewPing creates a new Ping service using the provided Connector.
func NewPing(conn Connector) Ping {
	return Ping{
		conn: conn,
	}
}

// Ping provides a connection check service for a given Connector.
func (pu Ping) Ping(ctx context.Context) error {
	if pu.conn == nil {
		log.Error().Msg("connecot is not initialized")
		return ErrConnNotInitialized
	}
	return pu.conn.Ping(ctx)
}
