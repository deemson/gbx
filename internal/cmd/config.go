package cmd

import (
	"errors"

	"github.com/deemson/gbx/internal/config"
	"github.com/spf13/cobra"
)

func configCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "config",
	}

	cmd.AddCommand(
		configWriteDefaultCmd(),
	)

	return cmd
}

func configWriteDefaultCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "write-default",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, _, err := config.Find()
			if err != nil && !errors.Is(err, config.ErrNotFound) {

			}
			return nil
		},
	}

	cmd.Flags().BoolP("force", "f", false, "force overwrite config")

	return cmd
}
