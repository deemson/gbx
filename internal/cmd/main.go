package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/adrg/xdg"
	"github.com/deemson/gbx/internal/tui"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func Main(version string) {
	cobra.EnableTraverseRunHooks = true

	cmd := &cobra.Command{
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("PersistentPreRunE")
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			zerolog.TimeFieldFormat = time.RFC3339Nano
			logPath, err := xdg.StateFile(fmt.Sprintf("gbx/gbx-%d.log", os.Getpid()))
			if err != nil {
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
			err = tui.Run(tui.WithDir(dir), tui.WithVersion(version), tui.WithLogPath(logPath))
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

	fmt.Println(cmd.Execute())
}
