package repository

import (
	"context"

	"avito/internal/domain"
)

type TeamRepository interface {
	Create(ctx context.Context, team *domain.Team) error
	Get(ctx context.Context, teamName string) (*domain.Team, error)
	GetByID(ctx context.Context, teamID int) (*domain.Team, error)
	Exists(ctx context.Context, teamName string) (bool, error)
	ExistsByID(ctx context.Context, teamID int) (bool, error)
	List(ctx context.Context) ([]*domain.Team, error)
	Count(ctx context.Context) (int, error)
}

type UserRepository interface {
	CreateOrUpdate(ctx context.Context, user *domain.User) error
	Get(ctx context.Context, userID string) (*domain.User, error)
	GetByTeamID(ctx context.Context, teamID int) ([]*domain.User, error)
	GetActiveByTeamID(ctx context.Context, teamID int) ([]*domain.User, error)
	GetActiveCandidatesForReview(ctx context.Context, teamID int, excludeUserIDs []string) ([]*domain.User, error)
	SetActive(ctx context.Context, userID string, isActive bool) error
	Exists(ctx context.Context, userID string) (bool, error)
	Delete(ctx context.Context, userID string) error
	List(ctx context.Context) ([]*domain.User, error)
	Count(ctx context.Context) (int, error)
}

type PullRequestRepository interface {
	Create(ctx context.Context, pr *domain.PullRequest) error
	Get(ctx context.Context, prID string) (*domain.PullRequest, error)
	Update(ctx context.Context, pr *domain.PullRequest) error
	Exists(ctx context.Context, prID string) (bool, error)
	GetReviewers(ctx context.Context, prID string) ([]string, error)
	SetReviewers(ctx context.Context, prID string, reviewerIDs []string) error
	AddReviewer(ctx context.Context, prID, userID string) error
	RemoveReviewer(ctx context.Context, prID, userID string) error
	ReplaceReviewer(ctx context.Context, prID, oldUserID, newUserID string) error
	GetByReviewer(ctx context.Context, userID string, openStatusID int16) ([]*domain.PullRequestShort, error)
	GetByAuthor(ctx context.Context, authorID string) ([]*domain.PullRequestShort, error)
	GetOpenPRs(ctx context.Context) ([]*domain.PullRequest, error)
	List(ctx context.Context) ([]*domain.PullRequest, error)
	Count(ctx context.Context) (int, error)
	Merge(ctx context.Context, prID string, mergedStatusID int16) (*domain.PullRequest, error)
}

type TaskRepository interface {
	CreateDeactivateTask(ctx context.Context, teamID int) error
	GetAndLockPendingTask(ctx context.Context) (*domain.BatchDeactivateTask, error)
	SetTaskStatus(ctx context.Context, taskID int, status string, errorMessage string) error
}
