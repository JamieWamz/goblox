package cli

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks",
	RunE: func(cmd *cobra.Command, args []string) error {
		status, _ := cmd.Flags().GetString("status")
		priority, _ := cmd.Flags().GetString("priority")

		db, err := getDB()
		if err != nil {
			return err
		}
		defer db.Close()

		tasks, err := db.ListTasks(context.Background(), status, priority)
		if err != nil {
			return err
		}

		if len(tasks) == 0 {
			fmt.Println("No tasks found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tDESCRIPTION\tPRIORITY\tSTATUS\tDUE DATE")
		for _, task := range tasks {
			dueDate := "-"
			if task.DueDate != nil {
				dueDate = task.DueDate.Format("2006-01-02")
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", task.ID[:8], task.Description, task.Priority, task.Status, dueDate)
		}
		w.Flush()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringP("status", "s", "", "Filter by status (pending, in_progress, done, archived)")
	listCmd.Flags().StringP("priority", "p", "", "Filter by priority (low, medium, high)")
}

var statusOptions = []string{"pending", "in_progress", "done", "archived"}

func selectStatus() (string, error) {
	var selected string
	prompt := &survey.Select{
		Message: "Select new status:",
		Options: statusOptions,
	}
	return selected, survey.AskOne(prompt, &selected)
}