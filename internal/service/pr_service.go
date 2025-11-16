package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand/v2"

	"avito/internal/domain"
	"avito/internal/repository"
	"avito/internal/repository/postgres"
)

type prService struct {
	db       *postgres.DB
	prRepo   repository.PullRequestRepository
	userRepo repository.UserRepository
	teamRepo repository.TeamRepository
}

func NewPRService(
	db *postgres.DB,
	prRepo repository.PullRequestRepository,
	userRepo repository.UserRepository,
	teamRepo repository.TeamRepository,
) PRService {
	return &prService{
		db:       db,
		prRepo:   prRepo,
		userRepo: userRepo,
		teamRepo: teamRepo,
	}
}

func (s *prService) CreatePR(ctx context.Context, prID, prName, authorID string) (*domain.PullRequest, error) {
	author, err := s.userRepo.Get(ctx, authorID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrAuthorNotFound
		}
		return nil, fmt.Errorf("failed to get author: %w", err)
	}
	team, err := s.teamRepo.GetByID(ctx, author.TeamID)
	if err != nil {
		if errors.Is(err, domain.ErrTeamNotFound) {
			return nil, fmt.Errorf("author's team not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get author's team: %w", err)
	}
	teamMembersDB, err := s.userRepo.GetByTeamID(ctx, team.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get team members: %w", err)
	}
	teamMembers := make([]*domain.TeamMember, 0, len(teamMembersDB))
	for _, u := range teamMembersDB {
		teamMembers = append(teamMembers, u.ToTeamMember())
	}
	teamDomain := domain.Team{
		ID:      team.ID,
		Name:    team.Name,
		Members: teamMembers,
	}
	reviewers := s.findReviewers(ctx, &teamDomain, authorID)
	reviewerIDs := make([]string, 0, len(reviewers))
	for _, r := range reviewers {
		reviewerIDs = append(reviewerIDs, r.UserID)
	}
	pr := &domain.PullRequest{
		PullRequestID:     prID,
		PullRequestName:   prName,
		AuthorID:          authorID,
		Status:            domain.PRStatusOpen,
		AssignedReviewers: reviewerIDs,
	}
	if err = pr.Validate(); err != nil {
		return nil, err
	}
	pr.PrepareForDB()
	err = s.db.WithTransaction(ctx, func(tx *sql.Tx) error {
		txPRRepo := postgres.NewPullRequestRepository(tx)
		if err = txPRRepo.Create(ctx, pr); err != nil {
			return fmt.Errorf("failed to create PR in repo: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	pr.SyncStatus()
	return pr, nil
}

func (s *prService) findReviewers(_ context.Context, team *domain.Team, authorID string) []*domain.TeamMember {
	candidates := team.GetActiveMembersExcluding(authorID)
	if len(candidates) == 0 {
		return []*domain.TeamMember{}
	}
	if len(candidates) <= 2 {
		return candidates
	}
	rand.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})
	return candidates[:2]
}

func (s *prService) MergePR(ctx context.Context, prID string) (*domain.PullRequest, error) {
	pr, err := s.prRepo.Get(ctx, prID)
	if err != nil {
		return nil, err
	}
	if pr.IsMerged() {
		return pr, nil
	}
	if err := s.prRepo.Merge(ctx, prID, domain.PRStatusIDMerged); err != nil {
		return nil, fmt.Errorf("failed to merge PR: %w", err)
	}
	return s.prRepo.Get(ctx, prID)
}

func (s *prService) ReassignReviewer(ctx context.Context, prID, oldReviewerID string) (*domain.PullRequest, string, error) {
	var newReviewerID string
	err := s.db.WithTransaction(ctx, func(tx *sql.Tx) error {
		txPRRepo := postgres.NewPullRequestRepository(tx)
		txUserRepo := postgres.NewUserRepository(tx)
		txTeamRepo := postgres.NewTeamRepository(tx)
		pr, err := txPRRepo.Get(ctx, prID)
		if err != nil {
			return err
		}
		if !pr.CanBeModified() {
			return domain.ErrPRMerged
		}
		if !pr.HasReviewer(oldReviewerID) {
			return domain.ErrNotAssigned
		}
		oldReviewer, err := txUserRepo.Get(ctx, oldReviewerID)
		if err != nil {
			return fmt.Errorf("failed to get old reviewer: %w", err)
		}
		team, err := txTeamRepo.GetByID(ctx, oldReviewer.TeamID)
		if err != nil {
			return fmt.Errorf("failed to get old reviewer's team: %w", err)
		}
		teamMembersDB, err := txUserRepo.GetByTeamID(ctx, team.ID)
		if err != nil {
			return fmt.Errorf("failed to get team members: %w", err)
		}
		teamMembers := make([]*domain.TeamMember, 0, len(teamMembersDB))
		for _, u := range teamMembersDB {
			teamMembers = append(teamMembers, u.ToTeamMember())
		}
		teamDomain := domain.Team{
			ID:      team.ID,
			Name:    team.Name,
			Members: teamMembers,
		}
		excludeIDs := make([]string, 0, len(pr.AssignedReviewers)+1)
		excludeIDs = append(excludeIDs, pr.AuthorID)
		excludeIDs = append(excludeIDs, pr.AssignedReviewers...)
		candidates := teamDomain.GetActiveMembersExcluding(excludeIDs...)
		if len(candidates) == 0 {
			return domain.ErrNoCandidate
		}
		newReviewer := candidates[rand.IntN(len(candidates))]
		newReviewerID = newReviewer.UserID
		if err := txPRRepo.ReplaceReviewer(ctx, prID, oldReviewerID, newReviewerID); err != nil {
			return fmt.Errorf("failed to replace reviewer in repo: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, "", err
	}
	updatedPR, err := s.prRepo.Get(ctx, prID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get updated PR: %w", err)
	}
	return updatedPR, newReviewerID, nil
}
