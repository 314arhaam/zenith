package cmd

import "github.com/spf13/cobra"

func newAddCommand(opts *options) *cobra.Command {
	return &cobra.Command{
		Use: "add SERVICE", Short: "Register a service", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name, err := serviceArg(args)
			if err != nil {
				return err
			}
			body, err := opts.client.Add(cmd.Context(), name)
			if err != nil {
				return err
			}
			_, err = cmd.OutOrStdout().Write(body)
			return err
		},
	}
}
