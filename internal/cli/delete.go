package cli

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
)

func newDeleteCommand(opts *options) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Permanently delete a task",
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

			if !force {
				confirmed := false
				prompt := &survey.Confirm{
					Message: fmt.Sprintf("Permanently delete %q?", task.Description),
					Default: false,
				}
				if err := survey.AskOne(prompt, &confirmed); err != nil {
					return fmt.Errorf("confirm deletion: %w", err)
				}
				if !confirmed {
					fmt.Fprintln(cmd.OutOrStdout(), "Cancelled.")
					return nil
				}
			}

			if err := db.DeleteTask(cmd.Context(), task.ID); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Deleted %s: %s\n", shortID(task.ID), task.Description)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "delete without confirmation")
	return cmd
}
