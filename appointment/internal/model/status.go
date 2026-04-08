package model

import "errors"

var ErrInvalidStatus = errors.New("status must be one of: new, in_progress, done")
var ErrStatusTransitionNotAllowed = errors.New("transitioning status from done back to new is not allowed")

type Status string

const (
	StatusNew        Status = "new"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

func (s Status) IsValid() bool {
	switch s {
	case StatusNew, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}
