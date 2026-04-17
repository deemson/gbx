package cmd

import (
	"os"

	"github.com/deemson/gbx/internal/lib/clilog"
	"github.com/deemson/gbx/internal/tui"
	"github.com/spf13/cobra"
)

func Main() {
	cobra.EnableTraverseRunHooks = true

	cmd := &cobra.Command{
		SilenceUsage:  true,
		SilenceErrors: true,
		Use:           "gbx",
		RunE: func(cmd *cobra.Command, args []string) error {
			return tui.Run()
		},
	}

	err := cmd.Execute()
	if err != nil {
		clilog.Error(err.Error())
		os.Exit(1)
	}
}
