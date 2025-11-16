package service

import (
	"context"
	"fmt"

	"avito/pkg/logger"
)

type userRepoForStats interface {
	Count(ctx context.Context) (int, error)
}

type teamRepoForStats interface {
	Count(ctx context.Context) (int, error)
}

type prRepoForStats interface {
	Count(ctx context.Context) (int, error)
}

type StatsService struct {
	userRepo userRepoForStats
	teamRepo teamRepoForStats
	prRepo   prRepoForStats
	logger   *logger.Logger
}

func NewStatsService(
	prRepo prRepoForStats,
	userRepo userRepoForStats,
	teamRepo teamRepoForStats,
	logger *logger.Logger,
) *StatsService {
	return &StatsService{
		userRepo: userRepo,
		teamRepo: teamRepo,
		prRepo:   prRepo,
		logger:   logger,
	}
}

type GlobalStats struct {
	TotalUsers int `json:"total_users"`
	TotalTeams int `json:"total_teams"`
	TotalPRs   int `json:"total_prs"`
}

func (s *StatsService) GetGlobalStats(ctx context.Context) (*GlobalStats, error) {
	users, err := s.userRepo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	teams, err := s.teamRepo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count teams: %w", err)
	}

	prs, err := s.prRepo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count prs: %w", err)
	}

	stats := &GlobalStats{
		TotalUsers: users,
		TotalTeams: teams,
		TotalPRs:   prs,
	}

	return stats, nil
}
