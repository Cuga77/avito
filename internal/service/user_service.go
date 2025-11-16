package service

import (
	"context"
	"fmt"

	"avito/internal/domain"
	"avito/pkg/logger"
)

type userRepoForUserService interface {
	Get(ctx context.Context, userID string) (*domain.User, error)
	SetActive(ctx context.Context, userID string, isActive bool) error
	Exists(ctx context.Context, userID string) (bool, error)
	GetByTeamID(ctx context.Context, teamID int) ([]*domain.User, error)
	GetActiveByTeamID(ctx context.Context, teamID int) ([]*domain.User, error)
}

type prRepoForUserService interface {
	GetByReviewer(ctx context.Context, userID string, openStatusID int16) ([]*domain.PullRequestShort, error)
}

type teamRepoForUserService interface {
	ExistsByID(ctx context.Context, teamID int) (bool, error)
}

type taskRepoForUserService interface {
	CreateDeactivateTask(ctx context.Context, teamID int) error
}

type prServiceForUserService interface {
	ReassignReviewer(ctx context.Context, prID, oldUserID string) (*domain.PullRequest, string, error)
}

type userService struct {
	userRepo  userRepoForUserService
	prRepo    prRepoForUserService
	prService prServiceForUserService
	teamRepo  teamRepoForUserService
	taskRepo  taskRepoForUserService
	logger    *logger.Logger
}

func NewUserService(
	userRepo userRepoForUserService,
	prRepo prRepoForUserService,
	prService prServiceForUserService,
	teamRepo teamRepoForUserService,
	taskRepo taskRepoForUserService,
	logger *logger.Logger,
) *userService {
	return &userService{
		userRepo:  userRepo,
		prRepo:    prRepo,
		prService: prService,
		teamRepo:  teamRepo,
		taskRepo:  taskRepo,
		logger:    logger,
	}
}

func (s *userService) SetIsActive(ctx context.Context, userID string, isActive bool) (*domain.User, error) {
	if userID == "" {
		return nil, domain.ErrInvalidInput
	}
	user, err := s.userRepo.Get(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user.IsActive == isActive {
		return user, nil
	}
	if err := s.userRepo.SetActive(ctx, userID, isActive); err != nil {
		return nil, fmt.Errorf("failed to set user active status: %w", err)
	}
	user.IsActive = isActive
	if !isActive {
		go s.triggerReassignment(userID)
	}
	return user, nil
}

func (s *userService) triggerReassignment(userID string) {
	ctx := context.Background()
	s.logger.Info("Запуск фонового переназначения для деактивированного пользователя", "userID", userID)

	openPRs, err := s.prRepo.GetByReviewer(ctx, userID, domain.PRStatusIDOpen)
	if err != nil {
		s.logger.Error("Не удалось получить PR для переназначения", "userID", userID, "error", err.Error())
		return
	}
	if len(openPRs) == 0 {
		s.logger.Info("У пользователя нет открытых PR для переназначения", "userID", userID)
		return
	}
	s.logger.Info(fmt.Sprintf("Найдено %d PR для переназначения", len(openPRs)), "userID", userID)

	for _, pr := range openPRs {
		_, _, err := s.prService.ReassignReviewer(ctx, pr.PullRequestID, userID)
		if err != nil {
			s.logger.Error("Не удалось переназначить PR",
				"prID", pr.PullRequestID,
				"userID", userID,
				"error", err.Error(),
			)
		} else {
			s.logger.Info("PR успешно переназначен", "prID", pr.PullRequestID)
		}
	}
	s.logger.Info("Фоновое переназначение завершено", "userID", userID)
}

func (s *userService) GetUser(ctx context.Context, userID string) (*domain.User, error) {
	if userID == "" {
		return nil, domain.ErrInvalidInput
	}
	user, err := s.userRepo.Get(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

func (s *userService) GetPRsByReviewer(ctx context.Context, userID string) ([]*domain.PullRequestShort, error) {
	if userID == "" {
		return nil, domain.ErrInvalidInput
	}
	if _, err := s.userRepo.Exists(ctx, userID); err != nil {
		return nil, domain.ErrUserNotFound
	}
	prs, err := s.prRepo.GetByReviewer(ctx, userID, domain.PRStatusIDOpen)
	if err != nil {
		return nil, fmt.Errorf("failed to get PRs by reviewer: %w", err)
	}
	if prs == nil {
		prs = []*domain.PullRequestShort{}
	}
	return prs, nil
}

func (s *userService) GetUsersByTeam(ctx context.Context, teamID int) ([]*domain.User, error) {
	users, err := s.userRepo.GetByTeamID(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to get users by team: %w", err)
	}
	return users, nil
}

func (s *userService) GetActiveUsersByTeam(ctx context.Context, teamID int) ([]*domain.User, error) {
	users, err := s.userRepo.GetActiveByTeamID(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active users by team: %w", err)
	}
	return users, nil
}

func (s *userService) ScheduleBatchDeactivate(ctx context.Context, teamID int) error {
	exists, err := s.teamRepo.ExistsByID(ctx, teamID)
	if err != nil {
		return fmt.Errorf("failed to check team existence: %w", err)
	}
	if !exists {
		return domain.ErrTeamNotFound
	}

	if err := s.taskRepo.CreateDeactivateTask(ctx, teamID); err != nil {
		return fmt.Errorf("failed to schedule task: %w", err)
	}

	s.logger.Info("Задача на массовую деактивацию успешно создана", "team_id", teamID)
	return nil
}
