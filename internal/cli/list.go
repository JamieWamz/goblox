package cli

import (
	"github.com/spf13/cobra"
)

func newListCommand(opts *options) *cobra.Command {
	var status string
	var priority string
	var format string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tasks",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			db, err := opts.database()
			if err != nil {
				return err
			}
			defer db.Close()

			tasks, err := db.ListTasks(cmd.Context(), status, priority)
			if err != nil {
				return err
			}
			return writeTasks(cmd.OutOrStdout(), tasks, format)
		},
	}

	cmd.Flags().StringVarP(&status, "status", "s", "", "filter by status: pending, in_progress, done, or archived")
	cmd.Flags().StringVarP(&priority, "priority", "p", "", "filter by priority: low, medium, or high")
	cmd.Flags().StringVarP(&format, "format", "f", formatTable, "output format: table, json, or csv")
	return cmd
}
