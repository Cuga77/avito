package handler

import (
	"log/slog"

	"avito/internal/service"
)

type Handler struct {
	teamService service.TeamService
	userService service.UserService
	prService   service.PRService
	logger      *slog.Logger
}

func NewHandler(
	teamService service.TeamService,
	userService service.UserService,
	prService service.PRService,
	logger *slog.Logger,
) *Handler {
	return &Handler{
		teamService: teamService,
		userService: userService,
		prService:   prService,
		logger:      logger,
	}
}
