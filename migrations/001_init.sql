-- Enable foreign keys
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS tasks (
    id TEXT PRIMARY KEY DEFAULT (hex(randomblob(16))),
    description TEXT NOT NULL CHECK(length(description) >= 3),
    priority TEXT NOT NULL CHECK(priority IN ("low", "medium", "high")),
    due_date DATETIME,
    status TEXT NOT NULL DEFAULT "pending" CHECK(status IN ("pending", "in_progress", "done", "archived")),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_priority ON tasks(priority);
CREATE INDEX idx_tasks_due_date ON tasks(due_date);

CREATE TRIGGER update_tasks_updated_at 
AFTER UPDATE ON tasks
BEGIN
    UPDATE tasks SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

INSERT INTO tasks (description, priority, due_date) VALUES
    ("Complete goblox MVP", "high", datetime("now", "+7 days")),
    ("Write comprehensive tests", "medium", datetime("now", "+14 days")),
    ("Create documentation", "low", datetime("now", "+21 days"));
