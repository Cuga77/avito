package service

import (
	"context"

	"avito/internal/domain"
)

type TeamService interface {
	CreateTeamWithMembers(ctx context.Context, team *domain.Team) (*domain.Team, error)
	GetTeamByName(ctx context.Context, teamName string) (*domain.Team, error)
}

// UserService интерфейс для работы с пользователями
type UserService interface {
	SetIsActive(ctx context.Context, userID string, isActive bool) (*domain.User, error)
	GetPRsByReviewer(ctx context.Context, userID string) ([]*domain.PullRequestShort, error)
	ScheduleBatchDeactivate(ctx context.Context, teamID int) error
	GetUser(ctx context.Context, userID string) (*domain.User, error)
}

var (
	_ TeamService = (*teamService)(nil)
	_ UserService = (*userService)(nil)
)
