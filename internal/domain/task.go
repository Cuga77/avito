package domain

import (
	"database/sql"
	"time"
)

const (
	TaskStatusPending    = "pending"
	TaskStatusProcessing = "processing"
	TaskStatusCompleted  = "completed"
	TaskStatusFailed     = "failed"
)

type BatchDeactivateTask struct {
	ID           int
	TeamID       int
	Status       string
	ErrorMessage sql.NullString
	CreatedAt    time.Time
	ProcessedAt  sql.NullTime
}
