package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"avito/pkg/response"
)

func (h *Handler) SetIsActive(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req SetIsActiveRequest
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

	user, err := h.userService.SetIsActive(ctx, req.UserID, req.IsActive)
	if err != nil {
		h.logger.Error("Failed to set user active status",
			"user_id", req.UserID,
			"is_active", req.IsActive,
			"error", err,
		)
		response.HandleError(w, err)
		return
	}

	h.logger.Info("User active status updated",
		"user_id", user.UserID,
		"is_active", user.IsActive,
	)

	resp := UserResponse{
		User: ToUserDTO(user),
	}

	response.OK(w, resp)
}

func (h *Handler) GetPRsByReviewer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		h.logger.Warn("Missing user_id parameter")
		response.BadRequest(w, "INVALID_INPUT", "user_id parameter is required")
		return
	}

	prs, err := h.userService.GetPRsByReviewer(ctx, userID)
	if err != nil {
		h.logger.Error("Failed to get PRs by reviewer",
			"user_id", userID,
			"error", err,
		)
		response.HandleError(w, err)
		return
	}

	h.logger.Info("PRs retrieved for reviewer",
		"user_id", userID,
		"count", len(prs),
	)

	resp := GetReviewResponse{
		UserID:       userID,
		PullRequests: ToPRShortDTOs(prs),
	}

	response.OK(w, resp)
}

func (h *Handler) BatchDeactivate(w http.ResponseWriter, r *http.Request) {
	var req BatchDeactivateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("Failed to decode batch deactivate request", "error", err)
		response.BadRequest(w, "INVALID_INPUT", "invalid request body")
		return
	}

	if err := req.Validate(); err != nil {
		h.logger.Warn("Invalid batch deactivate request", "error", err)
		response.HandleError(w, err)
		return
	}

	h.logger.Info("Batch deactivation request received", "team_id", req.TeamID)

	if err := h.userService.ScheduleBatchDeactivate(r.Context(), req.TeamID); err != nil {
		h.logger.Error("Failed to schedule batch deactivate", "error", err, "team_id", req.TeamID)
		response.HandleError(w, err)
		return
	}

	response.Accepted(w, map[string]string{
		"status":  "pending",
		"message": "Batch deactivation task has been scheduled.",
		"team_id": fmt.Sprintf("%d", req.TeamID),
	})
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID := r.PathValue("user_id")
	if userID == "" {
		h.logger.Warn("Missing user_id parameter")
		response.BadRequest(w, "INVALID_INPUT", "user_id is required")
		return
	}

	user, err := h.userService.GetUser(ctx, userID)
	if err != nil {
		h.logger.Error("Failed to get user",
			"user_id", userID,
			"error", err,
		)
		response.HandleError(w, err)
		return
	}

	h.logger.Info("User retrieved successfully",
		"user_id", user.UserID,
		"is_active", user.IsActive,
	)

	resp := UserResponse{
		User: ToUserDTO(user),
	}

	response.OK(w, resp)
}
