package handler

import (
	"encoding/json"
	"net/http"

	"avito/pkg/response"
)

func (h *Handler) CreatePR(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req CreatePRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("Failed to decode request body", "error", err)
		response.BadRequest(w, "INVALID_INPUT", "invalid request body")
		return
	}

	if err := req.Validate(); err != nil {
		h.logger.Warn("Invalid request data", "error", err)
		response.HandleError(w, err)
		return
	}

	pr, err := h.prService.CreatePR(ctx, req.PullRequestID, req.PullRequestName, req.AuthorID)
	if err != nil {
		h.logger.Error("Failed to create PR",
			"pr_id", req.PullRequestID,
			"author_id", req.AuthorID,
			"error", err,
		)
		response.HandleError(w, err)
		return
	}

	h.logger.Info("PR created successfully",
		"pr_id", pr.PullRequestID,
		"author_id", pr.AuthorID,
		"reviewers_count", len(pr.AssignedReviewers),
	)

	resp := PRResponse{
		PR: ToPRDTO(pr),
	}

	response.Created(w, resp)
}

func (h *Handler) MergePR(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req MergePRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("Failed to decode request body", "error", err)
		response.BadRequest(w, "INVALID_INPUT", "invalid request body")
		return
	}

	if err := req.Validate(); err != nil {
		h.logger.Warn("Invalid request data", "error", err)
		response.HandleError(w, err)
		return
	}

	pr, err := h.prService.MergePR(ctx, req.PullRequestID)
	if err != nil {
		h.logger.Error("Failed to merge PR",
			"pr_id", req.PullRequestID,
			"error", err,
		)
		response.HandleError(w, err)
		return
	}

	h.logger.Info("PR merged successfully",
		"pr_id", pr.PullRequestID,
		"status", pr.Status,
	)

	resp := PRResponse{
		PR: ToPRDTO(pr),
	}

	response.OK(w, resp)
}

func (h *Handler) ReassignReviewer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req ReassignReviewerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("Failed to decode request body", "error", err)
		response.BadRequest(w, "INVALID_INPUT", "invalid request body")
		return
	}

	if err := req.Validate(); err != nil {
		h.logger.Warn("Invalid request data", "error", err)
		response.HandleError(w, err)
		return
	}

	pr, newReviewerID, err := h.prService.ReassignReviewer(ctx, req.PullRequestID, req.OldUserID)
	if err != nil {
		h.logger.Error("Failed to reassign reviewer",
			"pr_id", req.PullRequestID,
			"old_user_id", req.OldUserID,
			"error", err,
		)
		response.HandleError(w, err)
		return
	}

	h.logger.Info("Reviewer reassigned successfully",
		"pr_id", pr.PullRequestID,
		"old_reviewer", req.OldUserID,
		"new_reviewer", newReviewerID,
	)

	resp := ReassignResponse{
		PR:         ToPRDTO(pr),
		ReplacedBy: newReviewerID,
	}

	response.OK(w, resp)
}
