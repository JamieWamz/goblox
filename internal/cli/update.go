package cli

import (
	"errors"
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/JamieWamz/goblox/internal/models"
	"github.com/spf13/cobra"
)

func newUpdateCommand(opts *options) *cobra.Command {
	var description string
	var priority string
	var status string
	var dueDate string
	var clearDue bool

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if clearDue && cmd.Flags().Changed("due") {
				return errors.New("--due and --clear-due cannot be used together")
			}

			db, err := opts.database()
			if err != nil {
				return err
			}
			defer db.Close()

			task, err := db.GetTask(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			changed := cmd.Flags().Changed("description") ||
				cmd.Flags().Changed("priority") ||
				cmd.Flags().Changed("status") ||
				cmd.Flags().Changed("due") || clearDue
			if !changed {
				if err := promptForTaskUpdate(opts, task); err != nil {
					return err
				}
			} else {
				if cmd.Flags().Changed("description") {
					task.Description = description
				}
				if cmd.Flags().Changed("priority") {
					task.Priority = models.Priority(priority)
				}
				if cmd.Flags().Changed("status") {
					task.Status = models.Status(status)
				}
				if cmd.Flags().Changed("due") {
					parsedDate, err := parseDueDate(dueDate)
					if err != nil {
						return err
					}
					task.DueDate = &parsedDate
				}
				if clearDue {
					task.DueDate = nil
				}
			}

			if err := task.Validate(); err != nil {
				return err
			}
			if err := db.UpdateTask(cmd.Context(), task); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Updated %s: %s\n", shortID(task.ID), task.Description)
			return nil
		},
	}

	cmd.Flags().StringVar(&description, "description", "", "new task description")
	cmd.Flags().StringVarP(&priority, "priority", "p", "", "new priority: low, medium, or high")
	cmd.Flags().StringVarP(&status, "status", "s", "", "new status: pending, in_progress, done, or archived")
	cmd.Flags().StringVarP(&dueDate, "due", "d", "", "new due date in YYYY-MM-DD format")
	cmd.Flags().BoolVar(&clearDue, "clear-due", false, "remove the due date")
	return cmd
}

func promptForTaskUpdate(opts *options, task *models.Task) error {
	description := task.Description
	priority := string(task.Priority)
	status := string(task.Status)
	dueDate := ""
	if task.DueDate != nil {
		dueDate = task.DueDate.Format(dateLayout)
	}

	questions := []*survey.Question{
		{Name: "description", Prompt: &survey.Input{Message: "Description:", Default: description}},
		{Name: "priority", Prompt: &survey.Select{Message: "Priority:", Options: priorityOptions, Default: priority}},
		{Name: "status", Prompt: &survey.Select{Message: "Status:", Options: statusOptions, Default: status}},
		{Name: "due_date", Prompt: &survey.Input{Message: "Due date (YYYY-MM-DD, empty to clear):", Default: dueDate}},
	}
	answers := struct {
		Description string `survey:"description"`
		Priority    string `survey:"priority"`
		Status      string `survey:"status"`
		DueDate     string `survey:"due_date"`
	}{}
	if err := opts.ask(questions, &answers); err != nil {
		return fmt.Errorf("read task updates: %w", err)
	}

	task.Description = answers.Description
	task.Priority = models.Priority(answers.Priority)
	task.Status = models.Status(answers.Status)
	if answers.DueDate == "" {
		task.DueDate = nil
		return nil
	}
	parsedDate, err := parseDueDate(answers.DueDate)
	if err != nil {
		return err
	}
	task.DueDate = &parsedDate
	return nil
}
