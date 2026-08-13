package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newExportCommand(opts *options) *cobra.Command {
	var status string
	var priority string
	var format string
	var outputPath string

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export tasks as JSON or CSV",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if format != formatJSON && format != formatCSV {
				return fmt.Errorf("invalid export format %q: use json or csv", format)
			}

			db, err := opts.database()
			if err != nil {
				return err
			}
			defer db.Close()

			tasks, err := db.ListTasks(cmd.Context(), status, priority)
			if err != nil {
				return err
			}

			if outputPath == "-" {
				return writeTasks(cmd.OutOrStdout(), tasks, format)
			}

			file, err := os.Create(outputPath)
			if err != nil {
				return fmt.Errorf("create export file: %w", err)
			}
			if err := writeTasks(file, tasks, format); err != nil {
				file.Close()
				return err
			}
			if err := file.Close(); err != nil {
				return fmt.Errorf("close export file: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Exported %d task(s) to %s\n", len(tasks), outputPath)
			return nil
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", formatJSON, "export format: json or csv")
	cmd.Flags().StringVarP(&outputPath, "output", "o", "-", "output file, or - for standard output")
	cmd.Flags().StringVarP(&status, "status", "s", "", "filter by status")
	cmd.Flags().StringVarP(&priority, "priority", "p", "", "filter by priority")
	return cmd
}
