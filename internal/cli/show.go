package cli

import (
	"github.com/JamieWamz/goblox/internal/models"
	"github.com/spf13/cobra"
)

func newShowCommand(opts *options) *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show one task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := opts.database()
			if err != nil {
				return err
			}
			defer db.Close()

			task, err := db.GetTask(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			return writeTasks(cmd.OutOrStdout(), []*models.Task{task}, format)
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", formatTable, "output format: table, json, or csv")
	return cmd
}
