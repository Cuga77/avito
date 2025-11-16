package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"avito/internal/domain"
)

type PullRequestRepository struct {
	db DBTX
}

func NewPullRequestRepository(db DBTX) *PullRequestRepository {
	return &PullRequestRepository{db: db}
}

func (r *PullRequestRepository) Create(ctx context.Context, pr *domain.PullRequest) error {
	pr.PrepareForDB()
	query := `
        INSERT INTO pull_requests (id, pull_request_name, author_id, status_id, created_at)
        VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP)
    `
	_, err := r.db.ExecContext(ctx, query,
		pr.PullRequestID,
		pr.PullRequestName,
		pr.AuthorID,
		pr.StatusID,
	)
	if err != nil {
		if isUniqueViolation(err, "pull_requests_pkey") {
			return domain.ErrPRExists
		}
		return fmt.Errorf("failed to create pull request: %w", err)
	}
	for _, reviewerID := range pr.AssignedReviewers {
		if err := r.AddReviewer(ctx, pr.PullRequestID, reviewerID); err != nil {
			return fmt.Errorf("failed to add reviewer %s: %w", reviewerID, err)
		}
	}
	return nil
}

func (r *PullRequestRepository) Get(ctx context.Context, prID string) (*domain.PullRequest, error) {
	query := `
        SELECT id, pull_request_name, author_id, status_id, created_at, merged_at
        FROM pull_requests
        WHERE id = $1
    `
	var pr domain.PullRequest
	var mergedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query, prID).Scan(
		&pr.PullRequestID,
		&pr.PullRequestName,
		&pr.AuthorID,
		&pr.StatusID,
		&pr.CreatedAt,
		&mergedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get pull request: %w", err)
	}
	if mergedAt.Valid {
		pr.MergedAt = &mergedAt.Time
	}
	pr.SyncStatus()
	reviewers, err := r.GetReviewers(ctx, prID)
	if err != nil {
		return nil, err
	}
	pr.AssignedReviewers = reviewers
	return &pr, nil
}

func (r *PullRequestRepository) Update(ctx context.Context, pr *domain.PullRequest) error {
	pr.PrepareForDB()
	query := `
        UPDATE pull_requests
        SET pull_request_name = $1,
            status_id = $2,
            merged_at = $3
        WHERE id = $4
    `
	result, err := r.db.ExecContext(ctx, query,
		pr.PullRequestName,
		pr.StatusID,
		pr.MergedAt,
		pr.PullRequestID,
	)
	if err != nil {
		return fmt.Errorf("failed to update pull request: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *PullRequestRepository) Exists(ctx context.Context, prID string) (bool, error) {
	query := `
        SELECT EXISTS(
            SELECT 1 FROM pull_requests WHERE id = $1
        )
    `
	var exists bool
	err := r.db.QueryRowContext(ctx, query, prID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check pull request existence: %w", err)
	}
	return exists, nil
}

func (r *PullRequestRepository) GetReviewers(ctx context.Context, prID string) ([]string, error) {
	query := `
        SELECT user_id
        FROM pr_reviewers
        WHERE pull_request_id = $1
        ORDER BY user_id
    `
	rows, err := r.db.QueryContext(ctx, query, prID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reviewers: %w", err)
	}
	defer rows.Close()
	reviewers := make([]string, 0)
	for rows.Next() {
		var reviewerID string
		if err := rows.Scan(&reviewerID); err != nil {
			return nil, fmt.Errorf("failed to scan reviewer: %w", err)
		}
		reviewers = append(reviewers, reviewerID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating reviewers: %w", err)
	}
	return reviewers, nil
}

func (r *PullRequestRepository) SetReviewers(ctx context.Context, prID string, reviewerIDs []string) error {
	deleteQuery := `
        DELETE FROM pr_reviewers
        WHERE pull_request_id = $1
    `
	if _, err := r.db.ExecContext(ctx, deleteQuery, prID); err != nil {
		return fmt.Errorf("failed to delete existing reviewers: %w", err)
	}
	if len(reviewerIDs) > 0 {
		insertQuery := `
            INSERT INTO pr_reviewers (pull_request_id, user_id)
            VALUES ($1, $2)
        `
		for _, reviewerID := range reviewerIDs {
			if _, err := r.db.ExecContext(ctx, insertQuery, prID, reviewerID); err != nil {
				return fmt.Errorf("failed to add reviewer %s: %w", reviewerID, err)
			}
		}
	}
	return nil
}

func (r *PullRequestRepository) AddReviewer(ctx context.Context, prID, userID string) error {
	query := `
        INSERT INTO pr_reviewers (pull_request_id, user_id)
        VALUES ($1, $2)
        ON CONFLICT (pull_request_id, user_id) DO NOTHING
    `
	_, err := r.db.ExecContext(ctx, query, prID, userID)
	if err != nil {
		return fmt.Errorf("failed to add reviewer: %w", err)
	}
	return nil
}

func (r *PullRequestRepository) RemoveReviewer(ctx context.Context, prID, userID string) error {
	query := `
        DELETE FROM pr_reviewers
        WHERE pull_request_id = $1 AND user_id = $2
    `
	result, err := r.db.ExecContext(ctx, query, prID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove reviewer: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return domain.ErrNotAssigned
	}
	return nil
}

func (r *PullRequestRepository) ReplaceReviewer(ctx context.Context, prID, oldUserID, newUserID string) error {
	deleteQuery := `
        DELETE FROM pr_reviewers
        WHERE pull_request_id = $1 AND user_id = $2
    `
	result, err := r.db.ExecContext(ctx, deleteQuery, prID, oldUserID)
	if err != nil {
		return fmt.Errorf("failed to remove old reviewer: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return domain.ErrNotAssigned
	}
	insertQuery := `
        INSERT INTO pr_reviewers (pull_request_id, user_id)
        VALUES ($1, $2)
    `
	_, err = r.db.ExecContext(ctx, insertQuery, prID, newUserID)
	if err != nil {
		return fmt.Errorf("failed to add new reviewer: %w", err)
	}
	return nil
}

func (r *PullRequestRepository) GetByReviewer(ctx context.Context, userID string, openStatusID int16) ([]*domain.PullRequestShort, error) {
	query := `
        SELECT p.id, p.pull_request_name, p.author_id, p.status_id
        FROM pull_requests p
        INNER JOIN pr_reviewers pr ON p.id = pr.pull_request_id
        WHERE pr.user_id = $1
        AND p.status_id = $2
    `
	rows, err := r.db.QueryContext(ctx, query, userID, openStatusID)
	if err != nil {
		return nil, fmt.Errorf("failed to get PRs by reviewer: %w", err)
	}
	defer rows.Close()
	var prs []*domain.PullRequestShort
	for rows.Next() {
		var pr domain.PullRequestShort
		var statusID int16
		if err := rows.Scan(
			&pr.PullRequestID,
			&pr.PullRequestName,
			&pr.AuthorID,
			&statusID,
		); err != nil {
			return nil, fmt.Errorf("failed to scan PR: %w", err)
		}
		pr.Status = domain.StatusIDToString(statusID)
		prs = append(prs, &pr)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating PRs: %w", err)
	}
	return prs, nil
}

func (r *PullRequestRepository) GetByAuthor(ctx context.Context, authorID string) ([]*domain.PullRequestShort, error) {
	query := `
        SELECT id, pull_request_name, author_id, status_id
        FROM pull_requests
        WHERE author_id = $1
        ORDER BY created_at DESC
    `
	rows, err := r.db.QueryContext(ctx, query, authorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get PRs by author: %w", err)
	}
	defer rows.Close()
	var prs []*domain.PullRequestShort
	for rows.Next() {
		var pr domain.PullRequestShort
		var statusID int16
		if err := rows.Scan(
			&pr.PullRequestID,
			&pr.PullRequestName,
			&pr.AuthorID,
			&statusID,
		); err != nil {
			return nil, fmt.Errorf("failed to scan PR: %w", err)
		}
		pr.Status = domain.StatusIDToString(statusID)
		prs = append(prs, &pr)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating PRs: %w", err)
	}
	return prs, nil
}

func (r *PullRequestRepository) GetOpenPRs(ctx context.Context) ([]*domain.PullRequest, error) {
	query := `
        SELECT id, pull_request_name, author_id, status_id, created_at, merged_at
        FROM pull_requests
        WHERE status_id = $1
        ORDER BY created_at DESC
    `
	rows, err := r.db.QueryContext(ctx, query, domain.PRStatusIDOpen)
	if err != nil {
		return nil, fmt.Errorf("failed to get open PRs: %w", err)
	}
	defer rows.Close()
	var prs []*domain.PullRequest
	for rows.Next() {
		var pr domain.PullRequest
		var mergedAt sql.NullTime
		if err := rows.Scan(
			&pr.PullRequestID,
			&pr.PullRequestName,
			&pr.AuthorID,
			&pr.StatusID,
			&pr.CreatedAt,
			&mergedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan PR: %w", err)
		}
		if mergedAt.Valid {
			pr.MergedAt = &mergedAt.Time
		}
		pr.SyncStatus()
		prs = append(prs, &pr)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating PRs: %w", err)
	}
	for _, pr := range prs {
		reviewers, err := r.GetReviewers(ctx, pr.PullRequestID)
		if err != nil {
			return nil, fmt.Errorf("failed to get reviewers for pr %s: %w", pr.PullRequestID, err)
		}
		pr.AssignedReviewers = reviewers
	}
	return prs, nil
}

func (r *PullRequestRepository) List(ctx context.Context) ([]*domain.PullRequest, error) {
	query := `
        SELECT id, pull_request_name, author_id, status_id, created_at, merged_at
        FROM pull_requests
        ORDER BY created_at DESC
    `
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list PRs: %w", err)
	}
	defer rows.Close()
	var prs []*domain.PullRequest
	for rows.Next() {
		var pr domain.PullRequest
		var mergedAt sql.NullTime
		if err := rows.Scan(
			&pr.PullRequestID,
			&pr.PullRequestName,
			&pr.AuthorID,
			&pr.StatusID,
			&pr.CreatedAt,
			&mergedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan PR: %w", err)
		}
		if mergedAt.Valid {
			pr.MergedAt = &mergedAt.Time
		}
		pr.SyncStatus()
		prs = append(prs, &pr)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating PRs: %w", err)
	}
	for _, pr := range prs {
		reviewers, err := r.GetReviewers(ctx, pr.PullRequestID)
		if err != nil {
			return nil, fmt.Errorf("failed to get reviewers for pr %s: %w", pr.PullRequestID, err)
		}
		pr.AssignedReviewers = reviewers
	}
	return prs, nil
}

func (r *PullRequestRepository) Count(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM pull_requests`
	if err := r.db.QueryRowContext(ctx, query).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count pull requests: %w", err)
	}
	return count, nil
}

func (r *PullRequestRepository) Merge(ctx context.Context, prID string, mergedStatusID int16) error {
	query := `
        UPDATE pull_requests
        SET
            status_id = $2,
            merged_at = COALESCE(merged_at, $3)
        WHERE
            id = $1
    `
	result, err := r.db.ExecContext(ctx, query, prID, mergedStatusID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to merge pull request: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return domain.ErrPRNotFound
	}
	return nil
}
