package cmd

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func getLogger(quiet, trace, verbose bool) zerolog.Logger {
	zerolog.DurationFieldInteger = true
	zerolog.DurationFieldUnit = time.Millisecond
	l := log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	switch {
	case quiet:
		return l.Level(zerolog.FatalLevel)
	case trace:
		return l.Level(zerolog.TraceLevel)
	case verbose:
		return l.Level(zerolog.DebugLevel)
	}

	return l.Level(zerolog.InfoLevel)
}
