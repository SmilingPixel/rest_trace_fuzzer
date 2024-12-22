package logger

import (
	"github.com/rs/zerolog"
)

func ConfigLogger() {
	// zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}
