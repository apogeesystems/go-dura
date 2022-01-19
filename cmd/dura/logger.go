package dura

import (
	"github.com/rs/zerolog"
)

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMicro
}

func SetGlobalLogLevel(level string) (err error) {
	var logLevel zerolog.Level
	if logLevel, err = zerolog.ParseLevel(level); err != nil {
		return
	}
	zerolog.SetGlobalLevel(logLevel)
	return
}
