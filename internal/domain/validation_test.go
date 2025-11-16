package domain_test

import (
	"testing"
	"time"

	"avito/internal/domain"
)

func TestUser_Validate(t *testing.T) {
	tests := []struct {
		name    string
		user    domain.User
		wantErr bool
	}{
		{
			name:    "Valid user",
			user:    domain.User{UserID: "valid_user", Username: "Test User", TeamID: 1},
			wantErr: false,
		},
		{
			name:    "Empty UserID",
			user:    domain.User{UserID: "", Username: "Test User", TeamID: 1},
			wantErr: true,
		},
		{
			name:    "Invalid UserID chars",
			user:    domain.User{UserID: "user@name", Username: "Test User", TeamID: 1},
			wantErr: true, // Regex check
		},
		{
			name:    "Empty Username",
			user:    domain.User{UserID: "valid_user", Username: "", TeamID: 1},
			wantErr: true,
		},
		{
			name:    "Invalid TeamID",
			user:    domain.User{UserID: "valid_user", Username: "Test User", TeamID: 0},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.user.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("User.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTeam_Validate(t *testing.T) {
	tests := []struct {
		name    string
		team    domain.Team
		wantErr bool
	}{
		{
			name:    "Valid team",
			team:    domain.Team{Name: "Valid Team", Members: []*domain.TeamMember{{UserID: "u1", Username: "n1"}}},
			wantErr: false,
		},
		{
			name:    "Empty Name",
			team:    domain.Team{Name: "", Members: []*domain.TeamMember{{UserID: "u1", Username: "n1"}}},
			wantErr: true,
		},
		{
			name:    "No Members",
			team:    domain.Team{Name: "Valid Team", Members: nil},
			wantErr: true,
		},
		{
			name:    "Invalid Member inside Team",
			team:    domain.Team{Name: "Valid Team", Members: []*domain.TeamMember{{UserID: "", Username: "n1"}}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.team.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Team.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPullRequest_Validate(t *testing.T) {
	validPR := domain.PullRequest{
		PullRequestID:     "pr-1",
		PullRequestName:   "Fix bug",
		AuthorID:          "user-1",
		Status:            domain.PRStatusOpen,
		AssignedReviewers: []string{},
	}

	tests := []struct {
		name    string
		modify  func(*domain.PullRequest)
		wantErr bool
	}{
		{
			name:    "Valid PR",
			modify:  func(pr *domain.PullRequest) {},
			wantErr: false,
		},
		{
			name:    "Empty ID",
			modify:  func(pr *domain.PullRequest) { pr.PullRequestID = "" },
			wantErr: true,
		},
		{
			name:    "Invalid Status",
			modify:  func(pr *domain.PullRequest) { pr.Status = "INVALID" },
			wantErr: true,
		},
		{
			name:    "Too many reviewers",
			modify:  func(pr *domain.PullRequest) { pr.AssignedReviewers = []string{"r1", "r2", "r3"} },
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := validPR // Copy
			tt.modify(&pr)
			if err := pr.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("PullRequest.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPullRequest_BusinessRules(t *testing.T) {
	pr := domain.PullRequest{StatusID: domain.PRStatusIDOpen}

	if !pr.IsOpen() {
		t.Error("Expected Open")
	}
	if pr.IsMerged() {
		t.Error("Expected Not Merged")
	}
	if !pr.CanBeModified() {
		t.Error("Expected Modifiable")
	}

	pr.Merge()

	if pr.IsOpen() {
		t.Error("Expected Not Open")
	}
	if !pr.IsMerged() {
		t.Error("Expected Merged")
	}
	if pr.CanBeModified() {
		t.Error("Expected Not Modifiable")
	}
	if pr.MergedAt == nil {
		t.Error("Expected MergedAt to be set")
	}
	if time.Since(*pr.MergedAt) > time.Second {
		t.Error("MergedAt is too old")
	}
}
