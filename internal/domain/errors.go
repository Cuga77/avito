package domain

import "errors"

var (
	ErrTeamExists     = errors.New("team already exists")
	ErrPRExists       = errors.New("pull request already exists")
	ErrPRMerged       = errors.New("pull request is already merged")
	ErrNotAssigned    = errors.New("user is not assigned as reviewer")
	ErrNoCandidate    = errors.New("no available candidate for assignment")
	ErrNotFound       = errors.New("resource not found")
	ErrInvalidInput   = errors.New("invalid input data")
	ErrAuthorNotFound = errors.New("author not found")
	ErrTeamNotFound   = errors.New("team not found")
	ErrUserNotFound   = errors.New("user not found")
	ErrPRNotFound     = errors.New("pull request not found")
)
