package cli

import (
	"fmt"

	"github.com/JamieWamz/goblox/internal/models"
	"github.com/spf13/cobra"
)

func newArchiveCommand(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:   "archive <id>",
		Short: "Archive a task without deleting it",
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
			task.Status = models.StatusArchived
			if err := db.UpdateTask(cmd.Context(), task); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Archived %s: %s\n", shortID(task.ID), task.Description)
			return nil
		},
	}
}
