package main

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		fmt.Println("$HOME is empty")
		os.Exit(1)
	}
	logFilePath := path.Join(homeDir, "gbx.log")
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		fmt.Println(fmt.Errorf("failed to open log file: %w", err).Error())
		os.Exit(1)
	}
	defer logFile.Close()
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	zerolog.TimeFieldFormat = time.RFC3339Nano
	log.Logger = zerolog.New(logFile).With().Timestamp().Logger()
	zerolog.DefaultContextLogger = &log.Logger
	log.Debug().Msg("sup")
	// if err := tui.Run(); err != nil {
	// 	os.Exit(1)
	// }
}
