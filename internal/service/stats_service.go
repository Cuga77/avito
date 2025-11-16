package service

import (
	"context"
	"fmt"

	"avito/internal/repository"
	"avito/pkg/logger"
)

type StatsService struct {
	userRepo repository.UserRepository
	teamRepo repository.TeamRepository
	prRepo   repository.PullRequestRepository
	logger   *logger.Logger
}

func NewStatsService(
	prRepo repository.PullRequestRepository,
	userRepo repository.UserRepository,
	teamRepo repository.TeamRepository,
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
