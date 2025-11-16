package service

import (
	"context"
	"database/sql"
	"fmt"

	"avito/internal/domain"
	"avito/internal/repository"

	"avito/internal/repository/postgres"
)

type teamService struct {
	db       *postgres.DB
	teamRepo repository.TeamRepository
	userRepo repository.UserRepository
}

func NewTeamService(
	db *postgres.DB,
	teamRepo repository.TeamRepository,
	userRepo repository.UserRepository,
) TeamService {
	return &teamService{
		db:       db,
		teamRepo: teamRepo,
		userRepo: userRepo,
	}
}

func (s *teamService) CreateTeamWithMembers(ctx context.Context, team *domain.Team) (*domain.Team, error) {
	if err := team.Validate(); err != nil {
		return nil, fmt.Errorf("invalid team: %w", err)
	}

	var createdTeam *domain.Team

	err := s.db.WithTransaction(ctx, func(tx *sql.Tx) error {
		txTeamRepo := postgres.NewTeamRepository(tx)
		txUserRepo := postgres.NewUserRepository(tx)

		exists, err := txTeamRepo.Exists(ctx, team.Name)
		if err != nil {
			return fmt.Errorf("failed to check team existence: %w", err)
		}
		if exists {
			return domain.ErrTeamExists
		}

		if err = txTeamRepo.Create(ctx, team); err != nil {
			return fmt.Errorf("failed to create team: %w", err)
		}
		teamID := team.ID

		for _, member := range team.Members {
			user := &domain.User{
				UserID:   member.UserID,
				Username: member.Username,
				TeamID:   teamID,
				IsActive: member.IsActive,
			}

			if err = txUserRepo.CreateOrUpdate(ctx, user); err != nil {
				return fmt.Errorf("failed to create/update user %s: %w", member.UserID, err)
			}
		}

		createdTeamDB, err := txTeamRepo.Get(ctx, team.Name)
		if err != nil {
			return fmt.Errorf("failed to get created team: %w", err)
		}

		createdTeam = createdTeamDB

		return nil
	})
	if err != nil {
		return nil, err
	}

	return createdTeam, nil
}

func (s *teamService) GetTeamByName(ctx context.Context, teamName string) (*domain.Team, error) {
	if teamName == "" {
		return nil, domain.ErrInvalidInput
	}

	team, err := s.teamRepo.Get(ctx, teamName)
	if err != nil {
		return nil, fmt.Errorf("failed to get team: %w", err)
	}

	return team, nil
}

func (s *teamService) TeamExists(ctx context.Context, teamName string) (bool, error) {
	if teamName == "" {
		return false, domain.ErrInvalidInput
	}

	exists, err := s.teamRepo.Exists(ctx, teamName)
	if err != nil {
		return false, fmt.Errorf("failed to check team existence: %w", err)
	}

	return exists, nil
}
