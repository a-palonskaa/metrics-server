package logger_test

import (
	"os"
	"testing"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"

	logger "github.com/a-palonskaa/metrics-server/pkg/logger"
)

func TestInitLogger_TimestampAndCaller(t *testing.T) {
	logFile := "test_logger_output.log"
	defer func() {
		if err := os.Remove(logFile); err != nil {
			log.Info().Err(err).Msg("failed to remove logFile")
		}
	}()

	logger.InitLogger(logFile)
	log.Info().Msg("testing logger format")

	data, err := os.ReadFile(logFile)
	require.NoError(t, err)

	logContent := string(data)

	require.Regexp(t, `\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}`, logContent, "timestamp missing or incorrect format")
	require.Regexp(t, `logger_test\.go:\d+`, logContent, "caller info missing")
}
