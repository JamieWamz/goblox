package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/JamieWamz/goblox/internal/models"
	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add [description]",
	Short: "Add a new task",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var description string
		if len(args) > 0 {
			description = args[0]
		} else {
			prompt := survey.Input{
				Message: "Task description:",
			}
			if err := survey.AskOne(&prompt, &description); err != nil {
				return err
			}
		}

		priority, _ := cmd.Flags().GetString("priority")
		dueDate, _ := cmd.Flags().GetString("due")

		task := &models.Task{
			Description: description,
			Priority:    models.Priority(priority),
			Status:      models.StatusPending,
		}

		if dueDate != "" {
			parsedDate, err := time.Parse("2006-01-02", dueDate)
			if err != nil {
				return fmt.Errorf("invalid due date format (use YYYY-MM-DD): %w", err)
			}
			task.DueDate = &parsedDate
		}

		if err := task.Validate(); err != nil {
			return err
		}

		db, err := getDB()
		if err != nil {
			return err
		}
		defer db.Close()

		if err := db.CreateTask(context.Background(), task); err != nil {
			return err
		}

		fmt.Printf("✓ Added task: %s\n", task.Description)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().StringP("priority", "p", "medium", "Task priority (low, medium, high)")
	addCmd.Flags().StringP("due", "d", "", "Due date (YYYY-MM-DD)")
}