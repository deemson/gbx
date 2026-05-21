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

	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}

	if err := tui.Run(tui.WithDir(dir)); err != nil {
		log.Fatal().Err(err).Msg("tui exited with error")
	}
}
