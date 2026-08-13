package cli

import (
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/JamieWamz/goblox/internal/models"
	"github.com/spf13/cobra"
)

func newAddCommand(opts *options) *cobra.Command {
	var priority string
	var dueDate string

	cmd := &cobra.Command{
		Use:   "add [description]",
		Short: "Add a task",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			description := strings.TrimSpace(strings.Join(args, " "))
			if description == "" {
				prompt := survey.Input{Message: "Task description:"}
				if err := survey.AskOne(&prompt, &description); err != nil {
					return fmt.Errorf("read description: %w", err)
				}
			}

			task := &models.Task{
				Description: description,
				Priority:    models.Priority(priority),
				Status:      models.StatusPending,
			}
			if dueDate != "" {
				parsedDate, err := parseDueDate(dueDate)
				if err != nil {
					return err
				}
				task.DueDate = &parsedDate
			}
			if err := task.Validate(); err != nil {
				return err
			}

			db, err := opts.database()
			if err != nil {
				return err
			}
			defer db.Close()

			if err := db.CreateTask(cmd.Context(), task); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Added %s: %s\n", shortID(task.ID), task.Description)
			return nil
		},
	}

	cmd.Flags().StringVarP(&priority, "priority", "p", string(models.PriorityMedium), "priority: low, medium, or high")
	cmd.Flags().StringVarP(&dueDate, "due", "d", "", "due date in YYYY-MM-DD format")
	return cmd
}
