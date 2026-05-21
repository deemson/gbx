package main

import (
	"os"
	"path"
	"time"

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
	// tui is supposed to be run here
}
