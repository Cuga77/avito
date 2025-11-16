package handler

import (
	"encoding/json"
	"net/http"

	"avito/internal/domain"
	"avito/pkg/response"
)

func (h *Handler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req CreateTeamRequest
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

	members := make([]*domain.TeamMember, 0, len(req.Members))
	for _, m := range req.Members {
		members = append(members, &domain.TeamMember{
			UserID:   m.UserID,
			Username: m.Username,
			IsActive: m.IsActive,
		})
	}

	team := &domain.Team{
		Name:    req.TeamName,
		Members: members,
	}

	createdTeam, err := h.teamService.CreateTeamWithMembers(ctx, team)
	if err != nil {
		h.logger.Error("Failed to create team",
			"team_name", req.TeamName,
			"error", err,
		)
		response.HandleError(w, err)
		return
	}

	h.logger.Info("Team created successfully",
		"team_id", createdTeam.ID,
		"team_name", createdTeam.Name,
		"members_count", len(createdTeam.Members),
	)

	resp := TeamResponse{
		Team: ToTeamDTO(createdTeam),
	}

	response.Created(w, resp)
}

func (h *Handler) GetTeam(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		h.logger.Warn("Missing team_name parameter")
		response.BadRequest(w, "INVALID_INPUT", "team_name parameter is required")
		return
	}

	team, err := h.teamService.GetTeamByName(ctx, teamName)
	if err != nil {
		h.logger.Error("Failed to get team",
			"team_name", teamName,
			"error", err,
		)
		response.HandleError(w, err)
		return
	}

	h.logger.Info("Team retrieved successfully",
		"team_id", team.ID,
		"team_name", team.Name,
		"members_count", len(team.Members),
	)

	resp := ToTeamDTO(team)

	response.OK(w, resp)
}
