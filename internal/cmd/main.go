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
	if err := rootCmd(version).Execute(); err != nil {
		clilog.Error(err.Error())
		os.Exit(1)
	}
}

func rootCmd(version string) *cobra.Command {
	cobra.EnableTraverseRunHooks = true

	cmd := &cobra.Command{
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(version, []string{"./"})
		},
	}

	cmd.AddCommand(
		configCmd(),
		openCmd(version),
	)
	return cmd
}

func run(version string, args []string) error {
	dirs, err := validateDirectories(args)
	if err != nil {
		return err
	}

	path, cfg, err := config.Find()
	if errors.Is(err, config.ErrNotFound) {
		cfg = config.Default()
	} else if err != nil {
		return fmt.Errorf("invalid config %s:\n%w", path, err)
	}

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

	err = tui.Run(tui.WithDirectories(dirs), tui.WithVersion(version), tui.WithLogPath(logPath), tui.WithConfig(cfg))
	if err != nil {
		log.Error().Err(err).Msg("tui exited with error")
	}
	_ = logFile.Close()
	if err != nil {
		crashPath := strings.TrimSuffix(logPath, ".log") + "-crash.log"
		if renameErr := os.Rename(logPath, crashPath); renameErr != nil {
			clilog.Errorf("could not save crash log as %s: %v\ncrash log left at %s", crashPath, renameErr, logPath)
		}
	} else {
		_ = os.Remove(logPath)
	}
	return err
}

func validateDirectories(args []string) ([]tui.Directory, error) {
	dirs := make([]tui.Directory, len(args))
	for i, arg := range args {
		info, err := os.Stat(arg)
		if err != nil {
			return nil, fmt.Errorf("open directory %q: %w", arg, err)
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("open directory %q: not a directory", arg)
		}
		if _, err := os.ReadDir(arg); err != nil {
			return nil, fmt.Errorf("open directory %q: %w", arg, err)
		}
		dirs[i] = tui.Directory{Label: arg, Path: arg}
	}
	return dirs, nil
}
