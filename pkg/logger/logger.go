// Package logger provides an initialization function for setting up
// a global zerolog-based structured logger that writes to a file or stderr.
package logger

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// InitLogger initializes the global zerolog logger.
//
// It attempts to open the given filename for appending logs.
// If the file cannot be opened, it falls back to standard error output.
//
// Logs are written with timestamps and caller information at the Info level.
func InitLogger(filename string) {
	ostream, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open log file(%s): %v", filename, err)
		ostream = os.Stderr
	}

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	zerolog.TimeFieldFormat = "2006-01-02 15:04:05"

	log.Logger = zerolog.New(ostream).With().Timestamp().Caller().Logger()
}
