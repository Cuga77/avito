package service

import (
	"context"

	"avito/internal/domain"
)

type TeamService interface {
	CreateTeamWithMembers(ctx context.Context, team *domain.Team) (*domain.Team, error)
	GetTeamByName(ctx context.Context, teamName string) (*domain.Team, error)
	TeamExists(ctx context.Context, teamName string) (bool, error)
}

type UserService interface {
	SetIsActive(ctx context.Context, userID string, isActive bool) (*domain.User, error)
	GetUser(ctx context.Context, userID string) (*domain.User, error)
	GetPRsByReviewer(ctx context.Context, userID string) ([]*domain.PullRequestShort, error)
	GetUsersByTeam(ctx context.Context, teamID int) ([]*domain.User, error)
	GetActiveUsersByTeam(ctx context.Context, teamID int) ([]*domain.User, error)
	ScheduleBatchDeactivate(ctx context.Context, teamID int) error
}

type PRService interface {
	CreatePR(ctx context.Context, prID, prName, authorID string) (*domain.PullRequest, error)
	MergePR(ctx context.Context, prID string) (*domain.PullRequest, error)
	ReassignReviewer(ctx context.Context, prID, oldReviewerID string) (*domain.PullRequest, string, error)
}
