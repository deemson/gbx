package cmd

import (
	"errors"
	"fmt"

	"github.com/deemson/gbx/internal/clilog"
	"github.com/deemson/gbx/internal/config"
	"github.com/spf13/cobra"
)

func configCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
	}

	cmd.AddCommand(
		configWriteDefaultCmd(),
	)

	return cmd
}

func configWriteDefaultCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "write-default",
		Short: "Write a default configuration with schema",
		RunE: func(cmd *cobra.Command, args []string) error {
			force, _ := cmd.Flags().GetBool("force")
			paths, err := config.WriteDefault(force)
			if _, ok := errors.AsType[*config.FileExistsError](err); ok {
				return fmt.Errorf("%w (use --force to overwrite)", err)
			}
			if err != nil {
				return err
			}
			for _, p := range paths {
				clilog.Infof("wrote %s", p)
			}
			return nil
		},
	}

	cmd.Flags().BoolP("force", "f", false, "force overwrite config")

	return cmd
}
