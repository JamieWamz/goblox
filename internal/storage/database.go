package storage

import (
    "context"
    "database/sql"
    "embed"
    "fmt"
    
    _ "github.com/mattn/go-sqlite3"
    "github.com/JamieWamz/goblox/internal/models"
)

//go:embed ../migrations/*.sql
var migrations embed.FS

type Database struct {
    db *sql.DB
}

func NewDatabase(dbPath string) (*Database, error) {
    db, err := sql.Open("sqlite3", dbPath+"?_journal=WAL&_busy_timeout=5000")
    if err != nil {
        return nil, fmt.Errorf("failed to open database: %w", err)
    }
    
    db.SetMaxOpenConns(1)
    db.SetMaxIdleConns(1)
    
    d := &Database{db: db}
    
    if err := d.migrate(context.Background()); err != nil {
        return nil, fmt.Errorf("failed to migrate: %w", err)
    }
    
    return d, nil
}

func (d *Database) migrate(ctx context.Context) error {
    migrationSQL, err := migrations.ReadFile("migrations/001_init.sql")
    if err != nil {
        return fmt.Errorf("failed to read migration: %w", err)
    }
    
    _, err = d.db.ExecContext(ctx, string(migrationSQL))
    if err != nil {
        return fmt.Errorf("failed to execute migration: %w", err)
    }
    
    return nil
}

func (d *Database) Close() error {
    return d.db.Close()
}

func (d *Database) GetDB() *sql.DB {
    return d.db
}

// CreateTask creates a new task in the database
func (d *Database) CreateTask(ctx context.Context, task *models.Task) error {
    _, err := d.db.ExecContext(ctx,
        "INSERT INTO tasks (description, priority, due_date, status) VALUES (?, ?, ?, ?)",
        task.Description, task.Priority, task.DueDate, task.Status,
    )
    if err != nil {
        return fmt.Errorf("failed to create task: %w", err)
    }
    return nil
}

// GetTask retrieves a task by ID
func (d *Database) GetTask(ctx context.Context, id string) (*models.Task, error) {
    var task models.Task
    var dueDate sql.NullTime
    
    err := d.db.QueryRowContext(ctx,
        "SELECT id, description, priority, due_date, status, created_at, updated_at FROM tasks WHERE id = ?",
        id,
    ).Scan(&task.ID, &task.Description, &task.Priority, &dueDate, &task.Status, &task.CreatedAt, &task.UpdatedAt)
    
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, fmt.Errorf("task not found: %s", id)
        }
        return nil, fmt.Errorf("failed to get task: %w", err)
    }
    
    if dueDate.Valid {
        task.DueDate = &dueDate.Time
    }
    
    return &task, nil
}

// ListTasks retrieves all tasks, optionally filtered by status and priority
func (d *Database) ListTasks(ctx context.Context, status, priority string) ([]*models.Task, error) {
    var rows *sql.Rows
    var err error
    
    if status != "" && priority != "" {
        rows, err = d.db.QueryContext(ctx,
            "SELECT id, description, priority, due_date, status, created_at, updated_at FROM tasks WHERE status = ? AND priority = ? ORDER BY created_at DESC",
            status, priority,
        )
    } else if status != "" {
        rows, err = d.db.QueryContext(ctx,
            "SELECT id, description, priority, due_date, status, created_at, updated_at FROM tasks WHERE status = ? ORDER BY created_at DESC",
            status,
        )
    } else if priority != "" {
        rows, err = d.db.QueryContext(ctx,
            "SELECT id, description, priority, due_date, status, created_at, updated_at FROM tasks WHERE priority = ? ORDER BY created_at DESC",
            priority,
        )
    } else {
        rows, err = d.db.QueryContext(ctx,
            "SELECT id, description, priority, due_date, status, created_at, updated_at FROM tasks ORDER BY created_at DESC",
        )
    }
    
    if err != nil {
        return nil, fmt.Errorf("failed to list tasks: %w", err)
    }
    defer rows.Close()
    
    var tasks []*models.Task
    for rows.Next() {
        var task models.Task
        var dueDate sql.NullTime
        
        err := rows.Scan(&task.ID, &task.Description, &task.Priority, &dueDate, &task.Status, &task.CreatedAt, &task.UpdatedAt)
        if err != nil {
            return nil, fmt.Errorf("failed to scan task: %w", err)
        }
        
        if dueDate.Valid {
            task.DueDate = &dueDate.Time
        }
        
        tasks = append(tasks, &task)
    }
    
    return tasks, nil
}

// UpdateTask updates an existing task
func (d *Database) UpdateTask(ctx context.Context, task *models.Task) error {
    var dueDate interface{}
    if task.DueDate != nil {
        dueDate = task.DueDate
    }
    
    _, err := d.db.ExecContext(ctx,
        "UPDATE tasks SET description = ?, priority = ?, due_date = ?, status = ? WHERE id = ?",
        task.Description, task.Priority, dueDate, task.Status, task.ID,
    )
    if err != nil {
        return fmt.Errorf("failed to update task: %w", err)
    }
    return nil
}

// DeleteTask removes a task from the database
func (d *Database) DeleteTask(ctx context.Context, id string) error {
    _, err := d.db.ExecContext(ctx, "DELETE FROM tasks WHERE id = ?", id)
    if err != nil {
        return fmt.Errorf("failed to delete task: %w", err)
    }
    return nil
}
```
