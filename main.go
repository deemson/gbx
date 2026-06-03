package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/adrg/xdg"
	"github.com/deemson/gbx/internal/tui"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// version is set at build time via -ldflags "-X main.version=v1.2.3"; empty in
// plain `go build`, where the TUI falls back to "dev".
var version string

func main() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	logPath, err := xdg.StateFile(fmt.Sprintf("gbx/gbx-%d.log", os.Getpid()))
	if err != nil {
		panic(err)
	}
	logFile, err := os.OpenFile(logPath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
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

	// The log is the post-mortem surface: it's removed on a clean exit, but kept
	// (renamed "-crash.log") when the TUI returns an error so the session that
	// failed can be inspected. Bubble Tea catches panics in the TUI and surfaces
	// them here as a non-nil error, so this covers crashes too.
	err = tui.Run(tui.WithDir(dir), tui.WithVersion(version))
	if err != nil {
		log.Error().Err(err).Msg("tui exited with error")
	}
	logFile.Close()
	if err != nil {
		os.Rename(logPath, strings.TrimSuffix(logPath, ".log")+"-crash.log")
	} else {
		os.Remove(logPath)
	}
}
