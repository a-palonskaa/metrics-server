package logger

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func InitLogger(filename string) {
	ostream, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open log file(%s): %v", filename, err)
		ostream = os.Stderr
	}

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	zerolog.TimeFieldFormat = "2006-01-02 15:04:05"

	log.Logger = zerolog.New(ostream).With().Timestamp().Caller().Logger()
}
