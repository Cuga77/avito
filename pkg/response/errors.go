package response

import (
	"errors"
	"net/http"

	"avito/internal/domain"
)

func HandleError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}

	switch {
	case errors.Is(err, domain.ErrTeamExists):
		BadRequest(w, "TEAM_EXISTS", "team already exists")

	case errors.Is(err, domain.ErrPRExists):
		Conflict(w, "PR_EXISTS", "pull request already exists")

	case errors.Is(err, domain.ErrPRMerged):
		Conflict(w, "PR_MERGED", "cannot modify merged pull request")

	case errors.Is(err, domain.ErrNotAssigned):
		Conflict(w, "NOT_ASSIGNED", "user is not assigned as reviewer")

	case errors.Is(err, domain.ErrNoCandidate):
		Conflict(w, "NO_CANDIDATE", "no available candidate for assignment")

	case errors.Is(err, domain.ErrNotFound):
		NotFound(w, "NOT_FOUND", "resource not found")

	case errors.Is(err, domain.ErrTeamNotFound):
		NotFound(w, "NOT_FOUND", "team not found")

	case errors.Is(err, domain.ErrUserNotFound):
		NotFound(w, "NOT_FOUND", "user not found")

	case errors.Is(err, domain.ErrAuthorNotFound):
		NotFound(w, "NOT_FOUND", "author not found")

	case errors.Is(err, domain.ErrInvalidInput):
		BadRequest(w, "INVALID_INPUT", "invalid input data")

	default:
		InternalError(w, "internal server error")
	}
}

func MapDomainErrorToHTTP(err error) int {
	switch {
	case errors.Is(err, domain.ErrTeamExists):
		return http.StatusBadRequest

	case errors.Is(err, domain.ErrPRExists):
		return http.StatusConflict

	case errors.Is(err, domain.ErrPRMerged):
		return http.StatusConflict

	case errors.Is(err, domain.ErrNotAssigned):
		return http.StatusConflict

	case errors.Is(err, domain.ErrNoCandidate):
		return http.StatusConflict

	case errors.Is(err, domain.ErrNotFound),
		errors.Is(err, domain.ErrTeamNotFound),
		errors.Is(err, domain.ErrUserNotFound),
		errors.Is(err, domain.ErrAuthorNotFound):
		return http.StatusNotFound

	case errors.Is(err, domain.ErrInvalidInput):
		return http.StatusBadRequest

	default:
		return http.StatusInternalServerError
	}
}

func MapDomainErrorToCode(err error) string {
	switch {
	case errors.Is(err, domain.ErrTeamExists):
		return "TEAM_EXISTS"

	case errors.Is(err, domain.ErrPRExists):
		return "PR_EXISTS"

	case errors.Is(err, domain.ErrPRMerged):
		return "PR_MERGED"

	case errors.Is(err, domain.ErrNotAssigned):
		return "NOT_ASSIGNED"

	case errors.Is(err, domain.ErrNoCandidate):
		return "NO_CANDIDATE"

	case errors.Is(err, domain.ErrNotFound),
		errors.Is(err, domain.ErrTeamNotFound),
		errors.Is(err, domain.ErrUserNotFound),
		errors.Is(err, domain.ErrAuthorNotFound):
		return "NOT_FOUND"

	case errors.Is(err, domain.ErrInvalidInput):
		return "INVALID_INPUT"

	default:
		return "INTERNAL_ERROR"
	}
}
