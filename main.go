package main

import (
	"os"
	"path"
	"time"

	"github.com/deemson/gbx/internal/tui"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	// logger := zerolog.New(zerolog.ConsoleWriter{
	// 	Out:        os.Stderr,
	// 	TimeFormat: time.TimeOnly,
	// }).With().Timestamp().Logger()
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		panic("empty home dir")
	}
	logFile, err := os.OpenFile(path.Join(homeDir, "gbx.log"), os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	log.Logger = zerolog.New(logFile).With().Timestamp().Logger()
	zerolog.DefaultContextLogger = &log.Logger
	if err := tui.Run(); err != nil {
		// logger.Fatal().Err(err).Msg("failure during tui.Run")
		panic(err)
	}
}
