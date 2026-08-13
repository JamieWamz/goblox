package cli

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/JamieWamz/goblox/internal/models"
	"github.com/JamieWamz/goblox/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func rejectPrompt(survey.Prompt, interface{}, ...survey.AskOpt) error {
	return errors.New("unexpected interactive prompt")
}

func rejectQuestions([]*survey.Question, interface{}, ...survey.AskOpt) error {
	return errors.New("unexpected interactive questions")
}

func executeCLI(t *testing.T, dbPath string, args ...string) (string, error) {
	t.Helper()
	return executeCLIWithOptions(t, &options{
		askOne: rejectPrompt,
		ask:    rejectQuestions,
	}, dbPath, args...)
}

func executeCLIWithOptions(t *testing.T, opts *options, dbPath string, args ...string) (string, error) {
	t.Helper()

	cmd := newRootCommand(opts)
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs(append([]string{"--db", dbPath}, args...))
	err := cmd.Execute()
	return output.String(), err
}

func addedID(t *testing.T, output string) string {
	t.Helper()
	fields := strings.Fields(output)
	require.GreaterOrEqual(t, len(fields), 2)
	return strings.TrimSuffix(fields[1], ":")
}

func TestTaskLifecycleCommands(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "tasks.db")

	output, err := executeCLI(t, dbPath, "list")
	require.NoError(t, err)
	assert.Equal(t, "No tasks found.\n", output)

	output, err = executeCLI(t, dbPath,
		"add", "Prepare", "release", "notes", "--priority", "high", "--due", "2030-01-02",
	)
	require.NoError(t, err)
	id := addedID(t, output)
	assert.Len(t, id, 8)
	assert.Contains(t, output, "Prepare release notes")

	output, err = executeCLI(t, dbPath, "add", "Publish release", "-p", "low")
	require.NoError(t, err)
	secondID := addedID(t, output)

	output, err = executeCLI(t, dbPath, "list", "--status", "pending", "--priority", "high")
	require.NoError(t, err)
	assert.Contains(t, output, "DUE DATE")
	assert.Contains(t, output, id)
	assert.NotContains(t, output, secondID)

	output, err = executeCLI(t, dbPath, "show", id, "--format", "json")
	require.NoError(t, err)
	var shown []*models.Task
	require.NoError(t, json.Unmarshal([]byte(output), &shown))
	require.Len(t, shown, 1)
	assert.Equal(t, "Prepare release notes", shown[0].Description)
	assert.Equal(t, "2030-01-02", shown[0].DueDate.Format(dateLayout))

	output, err = executeCLI(t, dbPath,
		"update", id,
		"--description", "Prepare final notes",
		"--priority", "medium",
		"--status", "done",
		"--clear-due",
	)
	require.NoError(t, err)
	assert.Contains(t, output, "Updated "+id)

	output, err = executeCLI(t, dbPath, "show", id, "--format", "csv")
	require.NoError(t, err)
	records, err := csv.NewReader(strings.NewReader(output)).ReadAll()
	require.NoError(t, err)
	require.Len(t, records, 2)
	assert.Equal(t, "Prepare final notes", records[1][1])
	assert.Equal(t, "done", records[1][3])
	assert.Empty(t, records[1][4])

	output, err = executeCLI(t, dbPath, "archive", id)
	require.NoError(t, err)
	assert.Contains(t, output, "Archived "+id)

	output, err = executeCLI(t, dbPath, "export", "--format", "json", "--status", "archived")
	require.NoError(t, err)
	var archived []*models.Task
	require.NoError(t, json.Unmarshal([]byte(output), &archived))
	require.Len(t, archived, 1)
	assert.Equal(t, models.StatusArchived, archived[0].Status)

	exportPath := filepath.Join(t.TempDir(), "tasks.csv")
	output, err = executeCLI(t, dbPath, "export", "--format", "csv", "--output", exportPath)
	require.NoError(t, err)
	assert.Contains(t, output, "Exported 2 task(s)")
	exported, err := os.ReadFile(exportPath)
	require.NoError(t, err)
	assert.Contains(t, string(exported), "Prepare final notes")
	assert.Contains(t, string(exported), "Publish release")

	output, err = executeCLI(t, dbPath, "delete", id, "--force")
	require.NoError(t, err)
	assert.Contains(t, output, "Deleted "+id)

	_, err = executeCLI(t, dbPath, "show", id)
	assert.ErrorIs(t, err, storage.ErrTaskNotFound)
}

func TestInteractiveCommandPaths(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "tasks.db")

	addOptions := &options{
		askOne: func(_ survey.Prompt, response interface{}, _ ...survey.AskOpt) error {
			*(response.(*string)) = "Prompt-created task"
			return nil
		},
		ask: rejectQuestions,
	}
	output, err := executeCLIWithOptions(t, addOptions, dbPath, "add")
	require.NoError(t, err)
	id := addedID(t, output)

	updateOptions := &options{
		askOne: rejectPrompt,
		ask: func(questions []*survey.Question, response interface{}, _ ...survey.AskOpt) error {
			assert.Len(t, questions, 4)
			value := reflect.ValueOf(response).Elem()
			value.FieldByName("Description").SetString("Interactively updated task")
			value.FieldByName("Priority").SetString("high")
			value.FieldByName("Status").SetString("in_progress")
			value.FieldByName("DueDate").SetString("2031-03-04")
			return nil
		},
	}
	output, err = executeCLIWithOptions(t, updateOptions, dbPath, "update", id)
	require.NoError(t, err)
	assert.Contains(t, output, "Interactively updated task")

	output, err = executeCLI(t, dbPath, "show", id, "--format", "json")
	require.NoError(t, err)
	assert.Contains(t, output, "2031-03-04")

	cancelOptions := &options{
		askOne: func(_ survey.Prompt, response interface{}, _ ...survey.AskOpt) error {
			*(response.(*bool)) = false
			return nil
		},
		ask: rejectQuestions,
	}
	output, err = executeCLIWithOptions(t, cancelOptions, dbPath, "delete", id)
	require.NoError(t, err)
	assert.Equal(t, "Cancelled.\n", output)

	confirmOptions := &options{
		askOne: func(_ survey.Prompt, response interface{}, _ ...survey.AskOpt) error {
			*(response.(*bool)) = true
			return nil
		},
		ask: rejectQuestions,
	}
	output, err = executeCLIWithOptions(t, confirmOptions, dbPath, "delete", id)
	require.NoError(t, err)
	assert.Contains(t, output, "Deleted "+id)
}

func TestCommandValidationErrors(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "tasks.db")

	_, err := executeCLI(t, dbPath, "add", "no")
	assert.ErrorContains(t, err, "description must be at least 3 characters")
	_, err = executeCLI(t, dbPath, "add", "Valid task", "--priority", "urgent")
	assert.ErrorContains(t, err, "invalid priority")
	_, err = executeCLI(t, dbPath, "add", "Valid task", "--due", "tomorrow")
	assert.ErrorContains(t, err, "use YYYY-MM-DD")

	output, err := executeCLI(t, dbPath, "add", "Task to update")
	require.NoError(t, err)
	id := addedID(t, output)

	_, err = executeCLI(t, dbPath, "update", id, "--due", "2030-01-01", "--clear-due")
	assert.ErrorContains(t, err, "cannot be used together")
	_, err = executeCLI(t, dbPath, "update", id, "--due", "invalid")
	assert.ErrorContains(t, err, "use YYYY-MM-DD")
	_, err = executeCLI(t, dbPath, "update", id, "--status", "unknown")
	assert.ErrorContains(t, err, "invalid status")

	_, err = executeCLI(t, dbPath, "list", "--status", "unknown")
	assert.ErrorContains(t, err, "invalid status")
	_, err = executeCLI(t, dbPath, "list", "--format", "yaml")
	assert.ErrorContains(t, err, "invalid output format")
	_, err = executeCLI(t, dbPath, "show")
	assert.Error(t, err)
	_, err = executeCLI(t, dbPath, "show", id, "--format", "yaml")
	assert.ErrorContains(t, err, "invalid output format")

	_, err = executeCLI(t, dbPath, "export", "--format", "yaml")
	assert.ErrorContains(t, err, "invalid export format")
	_, err = executeCLI(t, dbPath, "export", "--output", t.TempDir())
	assert.ErrorContains(t, err, "create export file")

	_, err = executeCLI(t, dbPath, "archive", "ffffffff")
	assert.ErrorIs(t, err, storage.ErrTaskNotFound)
	_, err = executeCLI(t, dbPath, "delete", "ffffffff", "--force")
	assert.ErrorIs(t, err, storage.ErrTaskNotFound)
	_, err = executeCLI(t, dbPath, "show", "%")
	assert.ErrorIs(t, err, storage.ErrInvalidTaskID)
}

func TestPromptErrorsAreReturned(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "tasks.db")
	promptErr := errors.New("terminal unavailable")
	opts := &options{
		askOne: func(survey.Prompt, interface{}, ...survey.AskOpt) error { return promptErr },
		ask:    rejectQuestions,
	}

	_, err := executeCLIWithOptions(t, opts, dbPath, "add")
	assert.ErrorIs(t, err, promptErr)

	output, err := executeCLI(t, dbPath, "add", "Existing task")
	require.NoError(t, err)
	id := addedID(t, output)

	_, err = executeCLIWithOptions(t, opts, dbPath, "delete", id)
	assert.ErrorIs(t, err, promptErr)

	updateOptions := &options{
		askOne: rejectPrompt,
		ask: func([]*survey.Question, interface{}, ...survey.AskOpt) error {
			return promptErr
		},
	}
	_, err = executeCLIWithOptions(t, updateOptions, dbPath, "update", id)
	assert.ErrorIs(t, err, promptErr)
}

func TestOutputHelpers(t *testing.T) {
	dueDate := time.Date(2030, time.May, 6, 0, 0, 0, 0, time.UTC)
	tasks := []*models.Task{{
		ID:          "1234567890abcdef",
		Description: "Export task",
		Priority:    models.PriorityHigh,
		Status:      models.StatusPending,
		DueDate:     &dueDate,
		CreatedAt:   dueDate,
		UpdatedAt:   dueDate,
	}}

	assert.Equal(t, "short", shortID("short"))
	assert.Equal(t, "12345678", shortID(tasks[0].ID))
	parsed, err := parseDueDate("2030-05-06")
	require.NoError(t, err)
	assert.Equal(t, dueDate, parsed)
	_, err = parseDueDate("May 6")
	assert.Error(t, err)

	var output bytes.Buffer
	require.NoError(t, writeTasks(&output, tasks, formatTable))
	assert.Contains(t, output.String(), "12345678")
	output.Reset()
	require.NoError(t, writeTasks(&output, []*models.Task{}, formatJSON))
	assert.Equal(t, "[]\n", output.String())
	output.Reset()
	require.NoError(t, writeTasks(&output, tasks, formatCSV))
	assert.Contains(t, output.String(), "2030-05-06")
	assert.Error(t, writeTasks(io.Discard, tasks, "xml"))
	assert.Error(t, writeTasks(failingWriter{}, tasks, formatJSON))
	assert.Error(t, writeTasks(failingWriter{}, tasks, formatTable))
	assert.Error(t, writeTasks(failingWriter{}, tasks, formatCSV))
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}
