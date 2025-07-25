package usecase

import (
	"context"
	"errors"

	"github.com/rs/zerolog/log"
)

var (
	ErrConnNotInitialized = errors.New("connector does not exist")
)

type Ping struct {
	conn Connector
}

func NewPing(conn Connector) Ping {
	return Ping{
		conn: conn,
	}
}

func (pu Ping) Ping(ctx context.Context) error {
	if pu.conn == nil {
		log.Error().Msg("connecot is not initialized")
		return ErrConnNotInitialized
	}
	return pu.conn.Ping(ctx)
}
