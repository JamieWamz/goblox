package cli

import (
	"context"
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		confirm := false
		prompt := &survey.Confirm{
			Message: "Are you sure you want to delete this task?",
			Default: false,
		}
		if err := survey.AskOne(prompt, &confirm); err != nil {
			return err
		}

		if !confirm {
			fmt.Println("Cancelled.")
			return nil
		}

		db, err := getDB()
		if err != nil {
			return err
		}
		defer db.Close()

		if err := db.DeleteTask(context.Background(), id); err != nil {
			return err
		}

		fmt.Printf("✓ Deleted task: %s\n", id[:8])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}