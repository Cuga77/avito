package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"avito/internal/domain"
)

type UserRepository struct {
	db DBTX
}

func NewUserRepository(db DBTX) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateOrUpdate(ctx context.Context, user *domain.User) error {
	query := `
        INSERT INTO users (id, username, team_id, is_active)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (id)
        DO UPDATE SET
            username = EXCLUDED.username,
            team_id = EXCLUDED.team_id,
            is_active = EXCLUDED.is_active
    `

	_, err := r.db.ExecContext(ctx, query,
		user.UserID,
		user.Username,
		user.TeamID,
		user.IsActive,
	)
	if err != nil {
		return fmt.Errorf("failed to create or update user: %w", err)
	}

	return nil
}

func (r *UserRepository) Get(ctx context.Context, userID string) (*domain.User, error) {
	query := `
        SELECT id, username, team_id, is_active
        FROM users
        WHERE id = $1
    `
	var user domain.User
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&user.UserID,
		&user.Username,
		&user.TeamID,
		&user.IsActive,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) GetByTeamID(ctx context.Context, teamID int) ([]*domain.User, error) {
	query := `
		SELECT id, username, team_id, is_active
		FROM users
		WHERE team_id = $1
		ORDER BY username
	`

	rows, err := r.db.QueryContext(ctx, query, teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to get users by team: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(
			&user.UserID,
			&user.Username,
			&user.TeamID,
			&user.IsActive,
		); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return users, nil
}

func (r *UserRepository) GetActiveByTeamID(ctx context.Context, teamID int) ([]*domain.User, error) {
	query := `
		SELECT id, username, team_id, is_active
		FROM users
		WHERE team_id = $1 AND is_active = true
		ORDER BY username
	`

	rows, err := r.db.QueryContext(ctx, query, teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active users by team: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(
			&user.UserID,
			&user.Username,
			&user.TeamID,
			&user.IsActive,
		); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return users, nil
}

func (r *UserRepository) GetActiveCandidatesForReview(ctx context.Context, teamID int, excludeUserIDs []string) ([]*domain.User, error) {
	if len(excludeUserIDs) == 0 {
		return r.GetActiveByTeamID(ctx, teamID)
	}

	query := `
		SELECT id, username, team_id, is_active
		FROM users
		WHERE team_id = $1
		  AND is_active = true
		  AND id != ALL($2)
		ORDER BY username
	`

	rows, err := r.db.QueryContext(ctx, query, teamID, excludeUserIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get active candidates: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(
			&user.UserID,
			&user.Username,
			&user.TeamID,
			&user.IsActive,
		); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return users, nil
}

func (r *UserRepository) SetActive(ctx context.Context, userID string, isActive bool) error {
	query := `
		UPDATE users
		SET is_active = $2
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, userID, isActive)
	if err != nil {
		return fmt.Errorf("failed to set user active status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

func (r *UserRepository) Exists(ctx context.Context, userID string) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS(
			SELECT 1 FROM users WHERE id = $1
		)
	`

	err := r.db.QueryRowContext(ctx, query, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}

	return exists, nil
}

func (r *UserRepository) Delete(ctx context.Context, userID string) error {
	query := `
		DELETE FROM users
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

func (r *UserRepository) List(ctx context.Context) ([]*domain.User, error) {
	query := `
		SELECT id, username, team_id, is_active
		FROM users
		ORDER BY username
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(
			&user.UserID,
			&user.Username,
			&user.TeamID,
			&user.IsActive,
		); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return users, nil
}

func (r *UserRepository) Count(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM users`
	if err := r.db.QueryRowContext(ctx, query).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}
	return count, nil
}
