package handler

import (
	"time"

	"avito/internal/domain"
)

type CreateTeamRequest struct {
	TeamName string           `json:"team_name"`
	Members  []*TeamMemberDTO `json:"members"`
}

type TeamMemberDTO struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

type TeamResponse struct {
	Team *TeamDTO `json:"team"`
}

type TeamDTO struct {
	ID      int              `json:"id"`
	Name    string           `json:"name"`
	Members []*TeamMemberDTO `json:"members"`
}

type SetIsActiveRequest struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

type UserResponse struct {
	User *UserDTO `json:"user"`
}

type UserDTO struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	TeamID   int    `json:"team_id"`
	IsActive bool   `json:"is_active"`
}

type CreatePRRequest struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
}

type MergePRRequest struct {
	PullRequestID string `json:"pull_request_id"`
}

type ReassignReviewerRequest struct {
	PullRequestID string `json:"pull_request_id"`
	OldUserID     string `json:"old_user_id"`
}

type PRResponse struct {
	PR *PRDTO `json:"pr"`
}

type ReassignResponse struct {
	PR         *PRDTO `json:"pr"`
	ReplacedBy string `json:"replaced_by"`
}

type PRDTO struct {
	PullRequestID     string     `json:"pull_request_id"`
	PullRequestName   string     `json:"pull_request_name"`
	AuthorID          string     `json:"author_id"`
	Status            string     `json:"status"`
	AssignedReviewers []string   `json:"assigned_reviewers"`
	CreatedAt         time.Time  `json:"createdAt"`
	MergedAt          *time.Time `json:"mergedAt,omitempty"`
}

type PRShortDTO struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
	Status          string `json:"status"`
}

type GetReviewResponse struct {
	UserID       string        `json:"user_id"`
	PullRequests []*PRShortDTO `json:"pull_requests"`
}

func ToTeamDTO(team *domain.Team) *TeamDTO {
	if team == nil {
		return nil
	}
	members := make([]*TeamMemberDTO, 0, len(team.Members))
	for _, m := range team.Members {
		members = append(members, &TeamMemberDTO{
			UserID:   m.UserID,
			Username: m.Username,
			IsActive: m.IsActive,
		})
	}
	return &TeamDTO{
		ID:      team.ID,
		Name:    team.Name,
		Members: members,
	}
}

func ToUserDTO(user *domain.User) *UserDTO {
	if user == nil {
		return nil
	}
	return &UserDTO{
		UserID:   user.UserID,
		Username: user.Username,
		TeamID:   user.TeamID,
		IsActive: user.IsActive,
	}
}

func ToPRDTO(pr *domain.PullRequest) *PRDTO {
	if pr == nil {
		return nil
	}
	return &PRDTO{
		PullRequestID:     pr.PullRequestID,
		PullRequestName:   pr.PullRequestName,
		AuthorID:          pr.AuthorID,
		Status:            string(pr.Status),
		AssignedReviewers: pr.AssignedReviewers,
		CreatedAt:         pr.CreatedAt,
		MergedAt:          pr.MergedAt,
	}
}

func ToPRShortDTOs(prs []*domain.PullRequestShort) []*PRShortDTO {
	if prs == nil {
		return []*PRShortDTO{}
	}
	dtos := make([]*PRShortDTO, 0, len(prs))
	for _, pr := range prs {
		dtos = append(dtos, &PRShortDTO{
			PullRequestID:   pr.PullRequestID,
			PullRequestName: pr.PullRequestName,
			AuthorID:        pr.AuthorID,
			Status:          string(pr.Status),
		})
	}
	return dtos
}

func (r *CreateTeamRequest) Validate() error {
	if r.TeamName == "" {
		return domain.ErrInvalidInput
	}
	if len(r.Members) == 0 {
		return domain.ErrInvalidInput
	}
	for _, member := range r.Members {
		if member.UserID == "" || member.Username == "" {
			return domain.ErrInvalidInput
		}
	}
	return nil
}

func (r *SetIsActiveRequest) Validate() error {
	if r.UserID == "" {
		return domain.ErrInvalidInput
	}
	return nil
}

func (r *CreatePRRequest) Validate() error {
	if r.PullRequestID == "" {
		return domain.ErrInvalidInput
	}
	if r.PullRequestName == "" {
		return domain.ErrInvalidInput
	}
	if r.AuthorID == "" {
		return domain.ErrInvalidInput
	}
	return nil
}

func (r *MergePRRequest) Validate() error {
	if r.PullRequestID == "" {
		return domain.ErrInvalidInput
	}
	return nil
}

func (r *ReassignReviewerRequest) Validate() error {
	if r.PullRequestID == "" {
		return domain.ErrInvalidInput
	}
	if r.OldUserID == "" {
		return domain.ErrInvalidInput
	}
	return nil
}

type BatchDeactivateRequest struct {
	TeamID int `json:"team_id"`
}

func (r *BatchDeactivateRequest) Validate() error {
	if r.TeamID <= 0 {
		return domain.ErrInvalidInput
	}
	return nil
}
