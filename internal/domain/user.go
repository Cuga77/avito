package domain

import (
	"fmt"
	"regexp"
)

var userIDRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

type User struct {
	UserID   string `json:"user_id" db:"id"`
	Username string `json:"username" db:"username"`
	TeamID   int    `json:"team_id" db:"team_id"`
	IsActive bool   `json:"is_active" db:"is_active"`
}

type TeamMember struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

func (u *User) ToTeamMember() *TeamMember {
	return &TeamMember{
		UserID:   u.UserID,
		Username: u.Username,
		IsActive: u.IsActive,
	}
}

func (u *User) Validate() error {
	if u.UserID == "" {
		return ErrInvalidInput
	}
	if !userIDRegex.MatchString(u.UserID) {
		return fmt.Errorf("invalid user ID format: %w", ErrInvalidInput)
	}
	if u.Username == "" {
		return ErrInvalidInput
	}
	if u.TeamID <= 0 {
		return ErrInvalidInput
	}
	return nil
}

func (tm *TeamMember) Validate() error {
	if tm.UserID == "" {
		return ErrInvalidInput
	}
	if !userIDRegex.MatchString(tm.UserID) {
		return fmt.Errorf("invalid team member user ID format: %w", ErrInvalidInput)
	}
	if tm.Username == "" {
		return ErrInvalidInput
	}
	return nil
}
