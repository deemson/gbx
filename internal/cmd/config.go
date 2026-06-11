package cmd

import (
	"github.com/deemson/gbx/internal/clilog"
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
			force, _ := cmd.Flags().GetBool("force")
			paths, err := config.WriteDefault(force)
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
