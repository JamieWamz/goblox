package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/JamieWamz/goblox/internal/models"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		db, err := getDB()
		if err != nil {
			return err
		}
		defer db.Close()

		task, err := db.GetTask(context.Background(), id)
		if err != nil {
			return err
		}

		description := task.Description
		priority := string(task.Priority)
		var dueDateStr string
		if task.DueDate != nil {
			dueDateStr = task.DueDate.Format("2006-01-02")
		}

		prompt := survey.Input{
			Message: "Description:",
			Default: task.Description,
		}
		if err := survey.AskOne(&prompt, &description); err != nil {
			return err
		}

		priorityPrompt := &survey.Select{
			Message: "Priority:",
			Options: []string{"low", "medium", "high"},
			Default: priority,
		}
		if err := survey.AskOne(priorityPrompt, &priority); err != nil {
			return err
		}

		duePrompt := survey.Input{
			Message: "Due date (YYYY-MM-DD, leave empty to clear):",
			Default: dueDateStr,
		}
		if err := survey.AskOne(&duePrompt, &dueDateStr); err != nil {
			return err
		}

		task.Description = description
		task.Priority = models.Priority(priority)

		if dueDateStr != "" {
			parsedDate, err := time.Parse("2006-01-02", dueDateStr)
			if err != nil {
				return fmt.Errorf("invalid due date format (use YYYY-MM-DD): %w", err)
			}
			task.DueDate = &parsedDate
		} else {
			task.DueDate = nil
		}

		if err := task.Validate(); err != nil {
			return err
		}

		if err := db.UpdateTask(context.Background(), task); err != nil {
			return err
		}

		fmt.Printf("✓ Updated task: %s\n", task.ID[:8])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
