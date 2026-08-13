package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPriorityIsValid(t *testing.T) {
	tests := []struct {
		name     string
		priority Priority
		expected bool
	}{
		{"low priority", PriorityLow, true},
		{"medium priority", PriorityMedium, true},
		{"high priority", PriorityHigh, true},
		{"invalid priority", Priority("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.priority.IsValid())
		})
	}
}

func TestStatusIsValid(t *testing.T) {
	tests := []struct {
		name     string
		status   Status
		expected bool
	}{
		{"pending status", StatusPending, true},
		{"in_progress status", StatusInProgress, true},
		{"done status", StatusDone, true},
		{"archived status", StatusArchived, true},
		{"invalid status", Status("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.status.IsValid())
		})
	}
}

func TestTaskValidate(t *testing.T) {
	tests := []struct {
		name        string
		task        *Task
		expectError bool
	}{
		{
			name: "valid task",
			task: &Task{
				Description: "Test task",
				Priority:    PriorityHigh,
				Status:      StatusPending,
			},
			expectError: false,
		},
		{
			name: "description too short",
			task: &Task{
				Description: "ab",
				Priority:    PriorityHigh,
				Status:      StatusPending,
			},
			expectError: true,
		},
		{
			name: "invalid priority",
			task: &Task{
				Description: "Test task",
				Priority:    Priority("invalid"),
				Status:      StatusPending,
			},
			expectError: true,
		},
		{
			name: "invalid status",
			task: &Task{
				Description: "Test task",
				Priority:    PriorityHigh,
				Status:      Status("invalid"),
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.task.Validate()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTaskWithDueDate(t *testing.T) {
	dueDate := time.Now().Add(24 * time.Hour)
	task := &Task{
		Description: "Task with due date",
		Priority:    PriorityMedium,
		Status:      StatusPending,
		DueDate:     &dueDate,
	}

	assert.NoError(t, task.Validate())
	assert.NotNil(t, task.DueDate)
}

func TestTaskValidateTrimsDescription(t *testing.T) {
	task := &Task{
		Description: "  Prepare release  ",
		Priority:    PriorityMedium,
		Status:      StatusPending,
	}

	assert.NoError(t, task.Validate())
	assert.Equal(t, "Prepare release", task.Description)
}
