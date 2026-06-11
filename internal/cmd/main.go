package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/deemson/gbx/internal/clilog"
	"github.com/deemson/gbx/internal/config"
	"github.com/deemson/gbx/internal/tui"
	"github.com/deemson/gbx/internal/xdg"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func Main(version string) {
	cobra.EnableTraverseRunHooks = true

	// cfg is loaded by the root PreRunE (TUI path only) and consumed by RunE.
	var cfg config.Config

	cmd := &cobra.Command{
		SilenceUsage:  true,
		SilenceErrors: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			_, loaded, err := config.Find()
			if errors.Is(err, config.ErrNotFound) {
				cfg = config.Default() // absent config → silent defaults
				return nil
			}
			if err != nil {
				return err // present but unreadable/invalid → hard fail
			}
			cfg = loaded
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			zerolog.TimeFieldFormat = time.RFC3339Nano
			logPath, err := xdg.StateFile(fmt.Sprintf("gbx/gbx-%d.log", os.Getpid()))
			if err != nil {
				return err
			}
			if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
				return err
			}
			logFile, err := os.OpenFile(logPath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
			if err != nil {
				return err
			}
			log.Logger = zerolog.New(logFile).With().Timestamp().Logger()
			zerolog.DefaultContextLogger = &log.Logger

			dir, err := os.Getwd()
			if err != nil {
				return err
			}
			if len(os.Args) > 1 {
				dir = os.Args[1]
			}

			// The log is the post-mortem surface: it's removed on a clean exit, but kept
			// (renamed "-crash.log") when the TUI returns an error so the session that
			// failed can be inspected. Bubble Tea catches panics in the TUI and surfaces
			// them here as a non-nil error, so this covers crashes too.
			err = tui.Run(tui.WithDir(dir), tui.WithVersion(version), tui.WithLogPath(logPath), tui.WithConfig(cfg))
			if err != nil {
				log.Error().Err(err).Msg("tui exited with error")
			}
			logFile.Close()
			if err != nil {
				os.Rename(logPath, strings.TrimSuffix(logPath, ".log")+"-crash.log")
			} else {
				os.Remove(logPath)
			}

			return err
		},
	}

	cmd.AddCommand(
		configCmd(),
	)

	if err := cmd.Execute(); err != nil {
		clilog.Error(err.Error())
		os.Exit(1)
	}
}
