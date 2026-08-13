package cli

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"text/tabwriter"
	"time"

	"github.com/JamieWamz/goblox/internal/models"
)

const (
	dateLayout  = "2006-01-02"
	formatTable = "table"
	formatJSON  = "json"
	formatCSV   = "csv"
)

var (
	priorityOptions = []string{"low", "medium", "high"}
	statusOptions   = []string{"pending", "in_progress", "done", "archived"}
)

func parseDueDate(value string) (time.Time, error) {
	parsed, err := time.Parse(dateLayout, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid due date %q: use YYYY-MM-DD", value)
	}
	return parsed, nil
}

func shortID(id string) string {
	if len(id) <= 8 {
		return id
	}
	return id[:8]
}

func writeTasks(w io.Writer, tasks []*models.Task, format string) error {
	switch format {
	case formatTable:
		return writeTaskTable(w, tasks)
	case formatJSON:
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(tasks); err != nil {
			return fmt.Errorf("encode JSON: %w", err)
		}
		return nil
	case formatCSV:
		return writeTaskCSV(w, tasks)
	default:
		return fmt.Errorf("invalid output format %q: use table, json, or csv", format)
	}
}

func writeTaskTable(w io.Writer, tasks []*models.Task) error {
	if len(tasks) == 0 {
		_, err := fmt.Fprintln(w, "No tasks found.")
		return err
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "ID\tDESCRIPTION\tPRIORITY\tSTATUS\tDUE DATE"); err != nil {
		return err
	}
	for _, task := range tasks {
		dueDate := "-"
		if task.DueDate != nil {
			dueDate = task.DueDate.Format(dateLayout)
		}
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
			shortID(task.ID), task.Description, task.Priority, task.Status, dueDate,
		); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return fmt.Errorf("flush table: %w", err)
	}
	return nil
}

func writeTaskCSV(w io.Writer, tasks []*models.Task) error {
	cw := csv.NewWriter(w)
	if err := cw.Write([]string{
		"id", "description", "priority", "status", "due_date", "created_at", "updated_at",
	}); err != nil {
		return fmt.Errorf("write CSV header: %w", err)
	}
	for _, task := range tasks {
		dueDate := ""
		if task.DueDate != nil {
			dueDate = task.DueDate.Format(dateLayout)
		}
		if err := cw.Write([]string{
			task.ID,
			task.Description,
			string(task.Priority),
			string(task.Status),
			dueDate,
			task.CreatedAt.Format(time.RFC3339),
			task.UpdatedAt.Format(time.RFC3339),
		}); err != nil {
			return fmt.Errorf("write CSV task: %w", err)
		}
	}
	cw.Flush()
	if err := cw.Error(); err != nil {
		return fmt.Errorf("flush CSV: %w", err)
	}
	return nil
}
