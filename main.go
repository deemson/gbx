package main

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

func main() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	logger := zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.TimeOnly,
	}).With().Timestamp().Logger()
	// homeDir := 
	// os.OpenFile()
}
