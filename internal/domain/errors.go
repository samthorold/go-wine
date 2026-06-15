package domain

import "errors"

var (
	// ErrNotFound is returned by repositories when an entity does not exist.
	ErrNotFound = errors.New("not found")

	// ErrInvalidRating is returned when a rating falls outside 1..5.
	ErrInvalidRating = errors.New("rating must be between 1 and 5")

	// ErrValidation is the umbrella for invariant violations when constructing
	// domain entities.
	ErrValidation = errors.New("validation failed")
)
