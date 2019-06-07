package cmd

import (
	"github.com/spf13/cobra"
)

func newRootCommand() *cobra.Command {
	rootCommand := &cobra.Command{
		Use:   "tezbird",
		Short: "A little birdy for the tezos network and twitter.",
	}

	rootCommand.AddCommand(
		newStartCommand(),
	)

	return rootCommand
}

func Execute() error {
	return newRootCommand().Execute()
}
