package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"avito/internal/domain"
)

type TaskRepository struct {
	db DBTX
}

func NewTaskRepository(db DBTX) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) CreateDeactivateTask(ctx context.Context, teamID int) error {
	query := `
        INSERT INTO batch_deactivate_tasks (team_id, status)
        VALUES ($1, $2)
    `
	_, err := r.db.ExecContext(ctx, query, teamID, domain.TaskStatusPending)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}
	return nil
}

func (r *TaskRepository) GetAndLockPendingTask(ctx context.Context) (*domain.BatchDeactivateTask, error) {
	query := `
        UPDATE batch_deactivate_tasks
        SET status = $1, processed_at = CURRENT_TIMESTAMP
        WHERE id = (
            SELECT id
            FROM batch_deactivate_tasks
            WHERE status = $2
            ORDER BY created_at
            FOR UPDATE SKIP LOCKED
            LIMIT 1
        )
        RETURNING id, team_id, status, error_message, created_at, processed_at
    `

	var task domain.BatchDeactivateTask
	err := r.db.QueryRowContext(ctx, query, domain.TaskStatusProcessing, domain.TaskStatusPending).Scan(
		&task.ID,
		&task.TeamID,
		&task.Status,
		&task.ErrorMessage,
		&task.CreatedAt,
		&task.ProcessedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get and lock task: %w", err)
	}
	return &task, nil
}

func (r *TaskRepository) SetTaskStatus(ctx context.Context, taskID int, status string, errorMessage string) error {
	var errMsg sql.NullString
	if errorMessage != "" {
		errMsg = sql.NullString{String: errorMessage, Valid: true}
	}

	query := `
        UPDATE batch_deactivate_tasks
        SET status = $1, error_message = $2
        WHERE id = $3
    `
	_, err := r.db.ExecContext(ctx, query, status, errMsg, taskID)
	if err != nil {
		return fmt.Errorf("failed to set task status: %w", err)
	}
	return nil
}
