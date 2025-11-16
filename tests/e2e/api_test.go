package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

var baseURL = getBaseURL()

func TestMain(m *testing.M) {
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get(baseURL + "/health")
		if err == nil {
			_ = resp.Body.Close()
			break
		}
		time.Sleep(1 * time.Second)
	}

	m.Run()
}

func TestFullLifecycle(t *testing.T) {
	runID := time.Now().UnixNano()
	teamName := fmt.Sprintf("e2e_team_%d", runID)
	authorID := fmt.Sprintf("e2e_author_%d", runID)
	reviewerID := fmt.Sprintf("e2e_reviewer_%d", runID)
	prID := fmt.Sprintf("e2e_pr_%d", runID)

	var createdPR *PRDTO

	t.Run("Create Team", func(t *testing.T) {
		teamReq := CreateTeamRequest{
			TeamName: teamName,
			Members: []*TeamMemberDTO{
				{UserID: authorID, Username: "E2E Author", IsActive: true},
				{UserID: reviewerID, Username: "E2E Reviewer", IsActive: true},
			},
		}
		resp, body := makeRequest(t, "POST", baseURL+"/team/add", teamReq)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("Expected status 201 Created, got %d. Body: %s", resp.StatusCode, string(body))
		}
	})

	t.Run("Create PR", func(t *testing.T) {
		prReq := CreatePRRequest{
			PullRequestID:   prID,
			PullRequestName: "E2E Test PR",
			AuthorID:        authorID,
		}
		resp, body := makeRequest(t, "POST", baseURL+"/pullRequest/create", prReq)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("Expected status 201 Created, got %d. Body: %s", resp.StatusCode, string(body))
		}

		var prResp struct {
			PR *PRDTO `json:"pr"`
		}
		if err := json.Unmarshal(body, &prResp); err != nil {
			t.Fatalf("Failed to parse PR response: %v. Body: %s", err, string(body))
		}
		createdPR = prResp.PR

		if createdPR.Status != "OPEN" {
			t.Errorf("Expected PR status 'OPEN', got '%s'", createdPR.Status)
		}
		if len(createdPR.AssignedReviewers) != 1 {
			t.Errorf("Expected 1 reviewer to be assigned, got %d", len(createdPR.AssignedReviewers))
		}
		if createdPR.AssignedReviewers[0] != reviewerID {
			t.Errorf("Expected reviewer '%s', got '%s'", reviewerID, createdPR.AssignedReviewers[0])
		}
	})

	t.Run("Merge PR", func(t *testing.T) {
		if createdPR == nil {
			t.Skip("Skipping Merge test, PR was not created")
		}

		mergeReq := MergePRRequest{PullRequestID: prID}

		resp, body := makeRequest(t, "POST", baseURL+"/pullRequest/merge", mergeReq)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected status 200 OK on merge, got %d. Body: %s", resp.StatusCode, string(body))
		}

		var prResp1 struct {
			PR *PRDTO `json:"pr"`
		}
		if err := json.Unmarshal(body, &prResp1); err != nil {
			t.Fatalf("Failed to parse merge response: %v. Body: %s", err, string(body))
		}
		if prResp1.PR.Status != "MERGED" {
			t.Errorf("Expected status 'MERGED', got '%s'", prResp1.PR.Status)
		}

		resp2, _ := makeRequest(t, "POST", baseURL+"/pullRequest/merge", mergeReq)
		defer resp2.Body.Close()

		if resp2.StatusCode != http.StatusOK {
			t.Fatalf("Expected status 200 OK on idempotent merge, got %d", resp2.StatusCode)
		}
	})

	t.Run("Reassign Merged PR Fails", func(t *testing.T) {
		if createdPR == nil {
			t.Skip("Skipping Reassign test, PR was not created")
		}

		reassignReq := ReassignReviewerRequest{
			PullRequestID: prID,
			OldUserID:     reviewerID,
		}

		resp, body := makeRequest(t, "POST", baseURL+"/pullRequest/reassign", reassignReq)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusConflict && resp.StatusCode != http.StatusForbidden {
			t.Fatalf("Expected status 409 or 403, got %d. Body: %s", resp.StatusCode, string(body))
		}

		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			t.Fatalf("Failed to parse error response: %v. Body: %s", err, string(body))
		}
		if errResp.Error.Code != "PR_MERGED" {
			t.Errorf("Expected error code 'PR_MERGED', got '%s'", errResp.Error.Code)
		}
	})
}

func TestBatchDeactivate(t *testing.T) {
	runID := time.Now().UnixNano()
	teamName := fmt.Sprintf("e2e_async_team_%d", runID)
	userA := fmt.Sprintf("e2e_async_user_a_%d", runID)
	userB := fmt.Sprintf("e2e_async_user_b_%d", runID)

	teamReq := CreateTeamRequest{
		TeamName: teamName,
		Members: []*TeamMemberDTO{
			{UserID: userA, Username: "E2E Async User A", IsActive: true},
			{UserID: userB, Username: "E2E Async User B", IsActive: true},
		},
	}

	resp, body := makeRequest(t, "POST", baseURL+"/team/add", teamReq)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Failed to create team for test: %s", string(body))
	}

	var teamResp struct {
		Team *TeamDTO `json:"team"`
	}
	if err := json.Unmarshal(body, &teamResp); err != nil {
		t.Fatalf("Failed to parse created team: %v", err)
	}
	teamID := teamResp.Team.ID

	deactivateReq := BatchDeactivateRequest{
		TeamID: teamID,
	}
	resp2, body2 := makeRequest(t, "POST", baseURL+"/users/batchDeactivate", deactivateReq)
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusAccepted {
		t.Fatalf("Expected status 202 Accepted, got %d. Body: %s", resp2.StatusCode, string(body2))
	}

	t.Logf("Waiting 15 seconds for background worker to process task (TeamID: %d)...", teamID)
	time.Sleep(15 * time.Second)

	t.Logf("Checking status for user %s...", userA)
	resp3, body3 := makeRequest(t, "GET", baseURL+"/users/"+userA, nil)
	defer resp3.Body.Close()

	if resp3.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200 OK for GET /users, got %d. Body: %s", resp3.StatusCode, string(body3))
	}

	var userResp UserResponse
	if err := json.Unmarshal(body3, &userResp); err != nil {
		t.Fatalf("Failed to parse User response: %v. Body: %s", err, string(body3))
	}

	if userResp.User.IsActive != false {
		t.Errorf("E2E test FAILED: User '%s' was not deactivated by the background worker. 'is_active' is still true.", userA)
	} else {
		t.Logf("E2E test SUCCESS: User '%s' was correctly deactivated.", userA)
	}
}
