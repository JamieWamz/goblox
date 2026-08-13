package storage

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/JamieWamz/goblox/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestDatabase(t *testing.T) *Database {
	t.Helper()

	db, err := NewDatabase(filepath.Join(t.TempDir(), "goblox.db"))
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})
	return db
}

func TestNewDatabaseRunsMigrationsOnce(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "goblox.db")

	db, err := NewDatabase(dbPath)
	require.NoError(t, err)

	var migrationCount int
	require.NoError(t, db.GetDB().QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&migrationCount))
	assert.Equal(t, 1, migrationCount)

	tasks, err := db.ListTasks(context.Background(), "", "")
	require.NoError(t, err)
	assert.Empty(t, tasks, "a new database must not contain sample tasks")
	require.NoError(t, db.Close())

	db, err = NewDatabase(dbPath)
	require.NoError(t, err)
	require.NoError(t, db.GetDB().QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&migrationCount))
	assert.Equal(t, 1, migrationCount)
	require.NoError(t, db.Close())
}

func TestNewDatabaseRejectsEmptyPath(t *testing.T) {
	db, err := NewDatabase("  ")

	assert.Nil(t, db)
	assert.EqualError(t, err, "database path cannot be empty")
}

func TestTaskCRUDAndFiltering(t *testing.T) {
	db := newTestDatabase(t)
	ctx := context.Background()
	dueDate := time.Date(2030, time.January, 2, 0, 0, 0, 0, time.UTC)
	task := &models.Task{
		Description: "  Ship the release  ",
		Priority:    models.PriorityHigh,
		Status:      models.StatusPending,
		DueDate:     &dueDate,
	}

	require.NoError(t, db.CreateTask(ctx, task))
	assert.Len(t, task.ID, 32)
	assert.Equal(t, "Ship the release", task.Description)
	assert.False(t, task.CreatedAt.IsZero())
	assert.False(t, task.UpdatedAt.IsZero())

	fetched, err := db.GetTask(ctx, task.ID[:8])
	require.NoError(t, err)
	assert.Equal(t, task.ID, fetched.ID)
	assert.Equal(t, dueDate, fetched.DueDate.UTC())

	second := &models.Task{
		Description: "Write release notes",
		Priority:    models.PriorityLow,
		Status:      models.StatusDone,
	}
	require.NoError(t, db.CreateTask(ctx, second))

	tasks, err := db.ListTasks(ctx, string(models.StatusPending), string(models.PriorityHigh))
	require.NoError(t, err)
	require.Len(t, tasks, 1)
	assert.Equal(t, task.ID, tasks[0].ID)

	_, err = db.ListTasks(ctx, "unknown", "")
	assert.EqualError(t, err, "invalid status: unknown")
	_, err = db.ListTasks(ctx, "", "urgent")
	assert.EqualError(t, err, "invalid priority: urgent")

	task.ID = task.ID[:8]
	task.Description = "Ship version 1.0"
	task.Status = models.StatusDone
	task.DueDate = nil
	require.NoError(t, db.UpdateTask(ctx, task))
	assert.Len(t, task.ID, 32)

	fetched, err = db.GetTask(ctx, task.ID)
	require.NoError(t, err)
	assert.Equal(t, "Ship version 1.0", fetched.Description)
	assert.Equal(t, models.StatusDone, fetched.Status)
	assert.Nil(t, fetched.DueDate)

	require.NoError(t, db.DeleteTask(ctx, task.ID[:8]))
	_, err = db.GetTask(ctx, task.ID)
	assert.ErrorIs(t, err, ErrTaskNotFound)
	assert.ErrorIs(t, db.DeleteTask(ctx, task.ID), ErrTaskNotFound)
}

func TestDatabaseValidatesWrites(t *testing.T) {
	db := newTestDatabase(t)
	ctx := context.Background()

	assert.EqualError(t, db.CreateTask(ctx, nil), "task cannot be nil")
	err := db.CreateTask(ctx, &models.Task{
		Description: "no",
		Priority:    models.PriorityMedium,
		Status:      models.StatusPending,
	})
	assert.ErrorContains(t, err, "description must be at least 3 characters")
	assert.EqualError(t, db.UpdateTask(ctx, nil), "task cannot be nil")
}

func TestTaskIDResolutionDetectsAmbiguousPrefixes(t *testing.T) {
	db := newTestDatabase(t)
	ctx := context.Background()
	for _, id := range []string{
		"deadbeef000000000000000000000001",
		"deadbeef000000000000000000000002",
	} {
		_, err := db.GetDB().ExecContext(ctx, `
			INSERT INTO tasks (id, description, priority, status)
			VALUES (?, 'Test task', 'medium', 'pending')`, id)
		require.NoError(t, err)
	}

	_, err := db.GetTask(ctx, "deadbeef")
	assert.ErrorIs(t, err, ErrAmbiguousTaskID)

	task, err := db.GetTask(ctx, "deadbeef000000000000000000000001")
	require.NoError(t, err)
	assert.Equal(t, "deadbeef000000000000000000000001", task.ID)
}

func TestTaskIDResolutionRejectsInvalidIDs(t *testing.T) {
	db := newTestDatabase(t)

	for _, id := range []string{"", "not-hex", "%", "123456789012345678901234567890123"} {
		_, err := db.GetTask(context.Background(), id)
		assert.ErrorIs(t, err, ErrInvalidTaskID)
	}
}

func TestDatabaseHonorsCancelledContext(t *testing.T) {
	db := newTestDatabase(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := db.CreateTask(ctx, &models.Task{
		Description: "Cancelled task",
		Priority:    models.PriorityMedium,
		Status:      models.StatusPending,
	})
	assert.Error(t, err)
	assert.False(t, errors.Is(err, ErrTaskNotFound))
}
