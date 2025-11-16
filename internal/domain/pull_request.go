package domain

import "time"

const (
	PRStatusIDOpen   int16 = 1
	PRStatusIDMerged int16 = 2
)

type PRStatus string

const (
	PRStatusOpen   PRStatus = "OPEN"
	PRStatusMerged PRStatus = "MERGED"
)

func (s PRStatus) IsValid() bool {
	return s == PRStatusOpen || s == PRStatusMerged
}

func (s PRStatus) ToStatusID() int16 {
	switch s {
	case PRStatusOpen:
		return PRStatusIDOpen
	case PRStatusMerged:
		return PRStatusIDMerged
	default:
		return PRStatusIDOpen
	}
}

func StatusIDToString(statusID int16) PRStatus {
	switch statusID {
	case PRStatusIDOpen:
		return PRStatusOpen
	case PRStatusIDMerged:
		return PRStatusMerged
	default:
		return PRStatusOpen
	}
}

type PullRequest struct {
	PullRequestID     string     `json:"pull_request_id" db:"id"`
	PullRequestName   string     `json:"pull_request_name" db:"pull_request_name"`
	AuthorID          string     `json:"author_id" db:"author_id"`
	StatusID          int16      `json:"-" db:"status_id"`
	Status            PRStatus   `json:"status" db:"-"`
	AssignedReviewers []string   `json:"assigned_reviewers,omitempty"`
	CreatedAt         time.Time  `json:"createdAt,omitempty" db:"created_at"`
	MergedAt          *time.Time `json:"mergedAt,omitempty" db:"merged_at"`
}

type PullRequestShort struct {
	PullRequestID   string   `json:"pull_request_id"`
	PullRequestName string   `json:"pull_request_name"`
	AuthorID        string   `json:"author_id"`
	Status          PRStatus `json:"status"`
}

func (pr *PullRequest) Validate() error {
	if pr.PullRequestID == "" {
		return ErrInvalidInput
	}
	if pr.PullRequestName == "" {
		return ErrInvalidInput
	}
	if pr.AuthorID == "" {
		return ErrInvalidInput
	}
	if !pr.Status.IsValid() {
		return ErrInvalidInput
	}
	if len(pr.AssignedReviewers) > 2 {
		return ErrInvalidInput
	}
	return nil
}

func (pr *PullRequest) IsMerged() bool {
	return pr.StatusID == PRStatusIDMerged
}

func (pr *PullRequest) IsOpen() bool {
	return pr.StatusID == PRStatusIDOpen
}

func (pr *PullRequest) CanBeModified() bool {
	return pr.IsOpen()
}

func (pr *PullRequest) HasReviewer(userID string) bool {
	for _, reviewerID := range pr.AssignedReviewers {
		if reviewerID == userID {
			return true
		}
	}
	return false
}

func (pr *PullRequest) IsAuthor(userID string) bool {
	return pr.AuthorID == userID
}

func (pr *PullRequest) ToShort() *PullRequestShort {
	return &PullRequestShort{
		PullRequestID:   pr.PullRequestID,
		PullRequestName: pr.PullRequestName,
		AuthorID:        pr.AuthorID,
		Status:          pr.Status,
	}
}

func (pr *PullRequest) Merge() {
	pr.StatusID = PRStatusIDMerged
	pr.Status = PRStatusMerged
	now := time.Now()
	pr.MergedAt = &now
}

func (pr *PullRequest) SyncStatus() {
	pr.Status = StatusIDToString(pr.StatusID)
}

func (pr *PullRequest) PrepareForDB() {
	pr.StatusID = pr.Status.ToStatusID()
}
