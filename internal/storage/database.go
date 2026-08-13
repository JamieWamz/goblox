package storage

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"strings"

	"github.com/JamieWamz/goblox/internal/models"
	_ "github.com/mattn/go-sqlite3"
)

var (
	ErrTaskNotFound    = errors.New("task not found")
	ErrAmbiguousTaskID = errors.New("task ID prefix is ambiguous")
	ErrInvalidTaskID   = errors.New("invalid task ID")
)

//go:embed migrations/*.sql
var migrations embed.FS

type Database struct {
	db *sql.DB
}

func NewDatabase(dbPath string) (*Database, error) {
	if strings.TrimSpace(dbPath) == "" {
		return nil, errors.New("database path cannot be empty")
	}

	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	d := &Database{db: db}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("connect to database: %w", err)
	}
	if err := d.migrate(context.Background()); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate database: %w", err)
	}

	return d, nil
}

func (d *Database) migrate(ctx context.Context) error {
	if _, err := d.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`); err != nil {
		return fmt.Errorf("create migration table: %w", err)
	}

	entries, err := fs.ReadDir(migrations, "migrations")
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		var applied bool
		err := d.db.QueryRowContext(ctx,
			"SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = ?)",
			entry.Name(),
		).Scan(&applied)
		if err != nil {
			return fmt.Errorf("check migration %s: %w", entry.Name(), err)
		}
		if applied {
			continue
		}

		migrationSQL, err := migrations.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return fmt.Errorf("read migration %s: %w", entry.Name(), err)
		}

		tx, err := d.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin migration %s: %w", entry.Name(), err)
		}
		if _, err := tx.ExecContext(ctx, string(migrationSQL)); err != nil {
			tx.Rollback()
			return fmt.Errorf("execute migration %s: %w", entry.Name(), err)
		}
		if _, err := tx.ExecContext(ctx,
			"INSERT INTO schema_migrations (version) VALUES (?)", entry.Name(),
		); err != nil {
			tx.Rollback()
			return fmt.Errorf("record migration %s: %w", entry.Name(), err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %s: %w", entry.Name(), err)
		}
	}

	return nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

func (d *Database) GetDB() *sql.DB {
	return d.db
}

// CreateTask validates and persists a new task, populating its generated fields.
func (d *Database) CreateTask(ctx context.Context, task *models.Task) error {
	if task == nil {
		return errors.New("task cannot be nil")
	}
	if err := task.Validate(); err != nil {
		return fmt.Errorf("validate task: %w", err)
	}

	return scanTask(d.db.QueryRowContext(ctx, `
		INSERT INTO tasks (description, priority, due_date, status)
		VALUES (?, ?, ?, ?)
		RETURNING id, description, priority, due_date, status, created_at, updated_at`,
		task.Description, task.Priority, task.DueDate, task.Status,
	), task)
}

// GetTask retrieves a task using either its full ID or a unique ID prefix.
func (d *Database) GetTask(ctx context.Context, id string) (*models.Task, error) {
	resolvedID, err := d.resolveTaskID(ctx, id)
	if err != nil {
		return nil, err
	}

	var task models.Task
	if err := scanTask(d.db.QueryRowContext(ctx, `
		SELECT id, description, priority, due_date, status, created_at, updated_at
		FROM tasks WHERE id = ?`, resolvedID), &task); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: %s", ErrTaskNotFound, id)
		}
		return nil, fmt.Errorf("get task: %w", err)
	}

	return &task, nil
}

// ListTasks retrieves tasks, optionally filtered by status and priority.
func (d *Database) ListTasks(ctx context.Context, status, priority string) ([]*models.Task, error) {
	if status != "" && !models.Status(status).IsValid() {
		return nil, fmt.Errorf("invalid status: %s", status)
	}
	if priority != "" && !models.Priority(priority).IsValid() {
		return nil, fmt.Errorf("invalid priority: %s", priority)
	}

	query := `SELECT id, description, priority, due_date, status, created_at, updated_at FROM tasks`
	var clauses []string
	var args []any
	if status != "" {
		clauses = append(clauses, "status = ?")
		args = append(args, status)
	}
	if priority != "" {
		clauses = append(clauses, "priority = ?")
		args = append(args, priority)
	}
	if len(clauses) > 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}
	query += " ORDER BY created_at DESC, id DESC"

	rows, err := d.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	tasks := make([]*models.Task, 0)
	for rows.Next() {
		var task models.Task
		if err := scanTask(rows, &task); err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		tasks = append(tasks, &task)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tasks: %w", err)
	}

	return tasks, nil
}

// UpdateTask validates and updates an existing task selected by full ID or prefix.
func (d *Database) UpdateTask(ctx context.Context, task *models.Task) error {
	if task == nil {
		return errors.New("task cannot be nil")
	}
	if err := task.Validate(); err != nil {
		return fmt.Errorf("validate task: %w", err)
	}

	resolvedID, err := d.resolveTaskID(ctx, task.ID)
	if err != nil {
		return err
	}

	result, err := d.db.ExecContext(ctx, `
		UPDATE tasks
		SET description = ?, priority = ?, due_date = ?, status = ?
		WHERE id = ?`,
		task.Description, task.Priority, task.DueDate, task.Status, resolvedID,
	)
	if err != nil {
		return fmt.Errorf("update task: %w", err)
	}
	if err := requireAffectedRow(result, task.ID); err != nil {
		return err
	}
	task.ID = resolvedID
	return nil
}

// DeleteTask removes a task selected by full ID or prefix.
func (d *Database) DeleteTask(ctx context.Context, id string) error {
	resolvedID, err := d.resolveTaskID(ctx, id)
	if err != nil {
		return err
	}

	result, err := d.db.ExecContext(ctx, "DELETE FROM tasks WHERE id = ?", resolvedID)
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}
	return requireAffectedRow(result, id)
}

func (d *Database) resolveTaskID(ctx context.Context, id string) (string, error) {
	id = strings.TrimSpace(id)
	if !isHexID(id) {
		return "", fmt.Errorf("%w: expected 1 to 32 hexadecimal characters", ErrInvalidTaskID)
	}

	rows, err := d.db.QueryContext(ctx, `
		SELECT id FROM tasks
		WHERE id = ? OR id LIKE ?
		ORDER BY CASE WHEN id = ? THEN 0 ELSE 1 END
		LIMIT 2`, id, id+"%", id)
	if err != nil {
		return "", fmt.Errorf("resolve task ID: %w", err)
	}
	defer rows.Close()

	var matches []string
	for rows.Next() {
		var match string
		if err := rows.Scan(&match); err != nil {
			return "", fmt.Errorf("scan task ID: %w", err)
		}
		matches = append(matches, match)
	}
	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("iterate task IDs: %w", err)
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("%w: %s", ErrTaskNotFound, id)
	}
	if matches[0] == id {
		return matches[0], nil
	}
	if len(matches) > 1 {
		return "", fmt.Errorf("%w: %s", ErrAmbiguousTaskID, id)
	}
	return matches[0], nil
}

func isHexID(id string) bool {
	if len(id) == 0 || len(id) > 32 {
		return false
	}
	for _, character := range id {
		if !((character >= '0' && character <= '9') ||
			(character >= 'a' && character <= 'f') ||
			(character >= 'A' && character <= 'F')) {
			return false
		}
	}
	return true
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanTask(scanner rowScanner, task *models.Task) error {
	var dueDate sql.NullTime
	if err := scanner.Scan(
		&task.ID,
		&task.Description,
		&task.Priority,
		&dueDate,
		&task.Status,
		&task.CreatedAt,
		&task.UpdatedAt,
	); err != nil {
		return err
	}
	if dueDate.Valid {
		task.DueDate = &dueDate.Time
	} else {
		task.DueDate = nil
	}
	return nil
}

func requireAffectedRow(result sql.Result, id string) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check affected rows: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("%w: %s", ErrTaskNotFound, id)
	}
	return nil
}
