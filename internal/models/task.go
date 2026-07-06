package models

import (
    "database/sql"
    "errors"
    "fmt"
    "strings"
    "time"
)

type Priority string

const (
    PriorityLow    Priority = "low"
    PriorityMedium Priority = "medium"
    PriorityHigh   Priority = "high"
)

func (p Priority) IsValid() bool {
    switch p {
    case PriorityLow, PriorityMedium, PriorityHigh:
        return true
    }
    return false
}

type Status string

const (
    StatusPending    Status = "pending"
    StatusInProgress Status = "in_progress"
    StatusDone       Status = "done"
    StatusArchived   Status = "archived"
)

func (s Status) IsValid() bool {
    switch s {
    case StatusPending, StatusInProgress, StatusDone, StatusArchived:
        return true
    }
    return false
}

type Task struct {
    ID          string     `json:"id"`
    Description string     `json:"description"`
    Priority    Priority   `json:"priority"`
    DueDate     *time.Time `json:"due_date,omitempty"`
    Status      Status     `json:"status"`
    CreatedAt   time.Time  `json:"created_at"`
    UpdatedAt   time.Time  `json:"updated_at"`
}

func (t *Task) Validate() error {
    if len(strings.TrimSpace(t.Description)) < 3 {
        return errors.New("description must be at least 3 characters")
    }
    
    if !t.Priority.IsValid() {
        return fmt.Errorf("invalid priority: %s", t.Priority)
    }
    
    if !t.Status.IsValid() {
        return fmt.Errorf("invalid status: %s", t.Status)
    }
    
    return nil
}

func (t *Task) ScanRow(rows *sql.Rows) error {
    var dueDate sql.NullTime
    err := rows.Scan(
        &t.ID,
        &t.Description,
        &t.Priority,
        &dueDate,
        &t.Status,
        &t.CreatedAt,
        &t.UpdatedAt,
    )
    if err != nil {
        return err
    }
    
    if dueDate.Valid {
        t.DueDate = &dueDate.Time
    }
    
    return nil
}
