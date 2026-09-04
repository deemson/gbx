package cmd

import "github.com/spf13/cobra"

func openCmd(version string) *cobra.Command {
	return &cobra.Command{
		Use:   "open <directory>...",
		Short: "Open one or more directories; calling this with a single argument './' is equivalent to bare gbx call",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(version, args)
		},
	}
}
