package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"

	"avito/internal/domain"
)

type TeamRepository struct {
	db DBTX
}

func NewTeamRepository(db DBTX) *TeamRepository {
	return &TeamRepository{db: db}
}

func (r *TeamRepository) Create(ctx context.Context, team *domain.Team) error {
	query := `
        INSERT INTO teams (name)
        VALUES ($1)
        RETURNING id
    `
	err := r.db.QueryRowContext(ctx, query, team.Name).Scan(&team.ID)
	if err != nil {
		if isUniqueViolation(err, "teams_name_key") {
			return domain.ErrTeamExists
		}
		return fmt.Errorf("failed to create team: %w", err)
	}
	return nil
}

func (r *TeamRepository) Get(ctx context.Context, teamName string) (*domain.Team, error) {
	query := `
        SELECT
            t.id,
            t.name,
            u.id,
            u.username,
            u.is_active
        FROM teams t
        LEFT JOIN users u ON t.id = u.team_id
        WHERE t.name = $1
    `
	rows, err := r.db.QueryContext(ctx, query, teamName)
	if err != nil {
		return nil, fmt.Errorf("failed to get team by name: %w", err)
	}
	defer rows.Close()

	var team *domain.Team
	members := []*domain.TeamMember{}

	for rows.Next() {
		var userID sql.NullString
		var userName sql.NullString
		var userIsActive sql.NullBool

		if team == nil {
			team = &domain.Team{}
		}

		if err := rows.Scan(
			&team.ID,
			&team.Name,
			&userID,
			&userName,
			&userIsActive,
		); err != nil {
			return nil, fmt.Errorf("failed to scan team or user: %w", err)
		}

		if userID.Valid {
			members = append(members, &domain.TeamMember{
				UserID:   userID.String,
				Username: userName.String,
				IsActive: userIsActive.Bool,
			})
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating team rows: %w", err)
	}

	if team == nil {
		return nil, domain.ErrTeamNotFound
	}

	team.Members = members
	return team, nil
}

func (r *TeamRepository) GetByID(ctx context.Context, teamID int) (*domain.Team, error) {
	var team domain.Team
	query := `
        SELECT id, name
        FROM teams
        WHERE id = $1
    `
	err := r.db.QueryRowContext(ctx, query, teamID).Scan(
		&team.ID,
		&team.Name,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrTeamNotFound
		}
		return nil, fmt.Errorf("failed to get team by id: %w", err)
	}
	return &team, nil
}

func (r *TeamRepository) Exists(ctx context.Context, teamName string) (bool, error) {
	var exists bool
	query := `
        SELECT EXISTS(
            SELECT 1 FROM teams WHERE name = $1
        )
    `
	err := r.db.QueryRowContext(ctx, query, teamName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check team existence: %w", err)
	}
	return exists, nil
}

func (r *TeamRepository) ExistsByID(ctx context.Context, teamID int) (bool, error) {
	var exists bool
	query := `
        SELECT EXISTS(
            SELECT 1 FROM teams WHERE id = $1
        )
    `
	err := r.db.QueryRowContext(ctx, query, teamID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check team existence by id: %w", err)
	}
	return exists, nil
}

func (r *TeamRepository) List(ctx context.Context) ([]*domain.Team, error) {
	query := `
        SELECT id, name
        FROM teams
        ORDER BY name
    `
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list teams: %w", err)
	}
	defer rows.Close()

	var teams []*domain.Team
	for rows.Next() {
		var team domain.Team
		if err := rows.Scan(&team.ID, &team.Name); err != nil {
			return nil, fmt.Errorf("failed to scan team: %w", err)
		}
		teams = append(teams, &team)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating teams: %w", err)
	}
	return teams, nil
}

func (r *TeamRepository) Count(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM teams`
	if err := r.db.QueryRowContext(ctx, query).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count teams: %w", err)
	}
	return count, nil
}

func isUniqueViolation(err error, constraintName string) bool {
	pqErr, ok := err.(*pq.Error)
	if !ok {
		return false
	}
	return pqErr.Code == "23505" && pqErr.Constraint == constraintName
}
