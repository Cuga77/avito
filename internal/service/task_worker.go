package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"avito/internal/domain"
	"avito/internal/repository"
	"avito/pkg/logger"
)

type TaskWorker struct {
	taskRepo  repository.TaskRepository
	userRepo  repository.UserRepository
	prService PRService
	logger    *logger.Logger
}

func NewTaskWorker(
	taskRepo repository.TaskRepository,
	userRepo repository.UserRepository,
	prService PRService,
	logger *logger.Logger,
) *TaskWorker {
	return &TaskWorker{
		taskRepo:  taskRepo,
		userRepo:  userRepo,
		prService: prService,
		logger:    logger,
	}
}

func (w *TaskWorker) Run(ctx context.Context) {
	w.logger.Info("Task worker started")
	ticker := time.NewTicker(5 * time.Second)

	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Task worker shutting down")
			return
		case <-ticker.C:
			w.processNextTask(ctx)
		}
	}
}

func (w *TaskWorker) processNextTask(ctx context.Context) {
	task, err := w.taskRepo.GetAndLockPendingTask(ctx)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return
		}
		w.logger.Error("Failed to get task", "error", err)
		return
	}

	w.logger.Info("Processing task", "task_id", task.ID, "team_id", task.TeamID)

	err = w.runDeactivation(ctx, task.TeamID)

	if err != nil {
		w.logger.Error("Task failed", "task_id", task.ID, "error", err)
		if statusErr := w.taskRepo.SetTaskStatus(ctx, task.ID, domain.TaskStatusFailed, err.Error()); statusErr != nil {
			w.logger.Error("Failed to set task status", "task_id", task.ID, "error", statusErr)
		}
	} else {
		w.logger.Info("Task completed", "task_id", task.ID)
		if statusErr := w.taskRepo.SetTaskStatus(ctx, task.ID, domain.TaskStatusCompleted, ""); statusErr != nil {
			w.logger.Error("Failed to set task status", "task_id", task.ID, "error", statusErr)
		}
	}
}

func (w *TaskWorker) runDeactivation(ctx context.Context, teamID int) error {
	users, err := w.userRepo.GetByTeamID(ctx, teamID)
	if err != nil {
		return fmt.Errorf("failed to get users for team %d: %w", teamID, err)
	}

	if len(users) == 0 {
		w.logger.Warn("No users found in team", "team_id", teamID)
		return nil
	}

	w.logger.Info(fmt.Sprintf("Found %d users to deactivate", len(users)), "team_id", teamID)

	for _, user := range users {
		if err := w.userRepo.SetActive(ctx, user.UserID, false); err != nil {
			w.logger.Error("Failed to deactivate user", "user_id", user.UserID, "error", err)
			continue
		}
		w.triggerReassignment(ctx, user.UserID)
	}

	w.logger.Info("Batch deactivation finished for team", "team_id", teamID)
	return nil
}

func (w *TaskWorker) triggerReassignment(_ context.Context, userID string) {
	w.logger.Info("Triggering reassignment", "user_id", userID)
}
