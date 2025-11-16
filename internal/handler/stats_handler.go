package handler

import (
	"net/http"

	"avito/internal/service"
	"avito/pkg/logger"
	"avito/pkg/response"
)

type StatsHandler struct {
	statsService *service.StatsService
	logger       *logger.Logger
}

func NewStatsHandler(statsService *service.StatsService, log *logger.Logger) *StatsHandler {
	return &StatsHandler{
		statsService: statsService,
		logger:       log,
	}
}

func (h *StatsHandler) GetGlobalStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	h.logger.Info("getting global stats")

	stats, err := h.statsService.GetGlobalStats(ctx)
	if err != nil {
		h.logger.Error("failed to get global stats", "error", err.Error())
		response.HandleError(w, err)
		return
	}

	response.OK(w, stats)
}

func (h *StatsHandler) GetTeamStats(w http.ResponseWriter, r *http.Request) {
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		response.BadRequest(w, "INVALID_INPUT", "team_name is required")
		return
	}
	h.logger.Info("getting team stats", "team_name", teamName)
	response.OK(w, "GetTeamStats not implemented")
}

func (h *StatsHandler) GetUserStats(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		response.BadRequest(w, "INVALID_INPUT", "user_id is required")
		return
	}
	h.logger.Info("getting user stats", "user_id", userID)
	response.OK(w, "GetUserStats not implemented")
}

func (h *StatsHandler) GetWorkloadStats(w http.ResponseWriter, r *http.Request) {
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		response.BadRequest(w, "INVALID_INPUT", "team_name is required")
		return
	}
	h.logger.Info("getting workload stats", "team_name", teamName)
	response.OK(w, "GetWorkloadStats not implemented")
}

func (h *StatsHandler) GetHealthStats(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("getting health stats")
	response.OK(w, "GetHealthStats not implemented")
}
