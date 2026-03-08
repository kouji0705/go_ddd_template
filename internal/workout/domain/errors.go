package domain

import "errors"

// ─── ドメインエラー ────────────────────────────────────────────────────────────

var (
	ErrWorkoutNotFound     = errors.New("workout not found")
	ErrEmptyName           = errors.New("workout name must not be empty")
	ErrNonPositiveCalorie  = errors.New("calories must be a positive number")
	ErrNonPositiveDuration = errors.New("duration must be a positive number")
	ErrInvalidWorkoutID    = errors.New("invalid workout id")
)
