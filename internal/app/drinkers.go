package app

import (
	"context"

	"go-wine/internal/domain"
)

// CreateDrinkerCommand is the input to adding a Drinker.
type CreateDrinkerCommand struct {
	Name string
}

// CreateDrinkerHandler is a command-side use case: it goes through the domain so
// the non-empty-name invariant is enforced in one place, then persists.
type CreateDrinkerHandler struct {
	drinkers domain.DrinkerRepository
}

func NewCreateDrinkerHandler(d domain.DrinkerRepository) *CreateDrinkerHandler {
	return &CreateDrinkerHandler{drinkers: d}
}

// Handle constructs a Drinker through the domain and saves it, returning the new ID.
func (h *CreateDrinkerHandler) Handle(ctx context.Context, cmd CreateDrinkerCommand) (domain.ID, error) {
	d, err := domain.NewDrinker(cmd.Name)
	if err != nil {
		return "", err
	}
	if err := h.drinkers.Save(ctx, d); err != nil {
		return "", err
	}
	return d.ID, nil
}

// RenameDrinkerCommand is the input to renaming an existing Drinker.
type RenameDrinkerCommand struct {
	ID   domain.ID
	Name string
}

// RenameDrinkerHandler loads an existing Drinker, renames it through the domain
// (which holds the non-empty-name invariant), and persists the change.
type RenameDrinkerHandler struct {
	drinkers domain.DrinkerRepository
}

func NewRenameDrinkerHandler(d domain.DrinkerRepository) *RenameDrinkerHandler {
	return &RenameDrinkerHandler{drinkers: d}
}

// Handle renames the Drinker, returning ErrNotFound for an unknown ID and
// ErrValidation for an empty name (in which case nothing is persisted).
func (h *RenameDrinkerHandler) Handle(ctx context.Context, cmd RenameDrinkerCommand) error {
	d, err := h.drinkers.Get(ctx, cmd.ID)
	if err != nil {
		return err
	}
	if err := d.Rename(cmd.Name); err != nil {
		return err
	}
	return h.drinkers.Save(ctx, d)
}
