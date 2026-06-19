package app

import (
	"context"

	"go-wine/internal/domain"
)

// CompositionPartInput is one Variety's share as submitted to the edit command.
// Proportion is a plain int percentage; the domain turns the set into a
// validated Composition.
type CompositionPartInput struct {
	VarietyID  domain.ID
	Proportion int
}

// EditCompositionCommand is the input to setting a Wine's Composition.
type EditCompositionCommand struct {
	WineID domain.ID
	Parts  []CompositionPartInput
}

// EditCompositionHandler is a command-side use case. It goes through the Wine
// aggregate because Composition carries the real invariant (≥1 Variety,
// proportions summing to ~100%). It also checks the referenced Wine and every
// Variety exist before persisting.
type EditCompositionHandler struct {
	wines     domain.WineRepository
	varieties domain.VarietyRepository
}

func NewEditCompositionHandler(w domain.WineRepository, v domain.VarietyRepository) *EditCompositionHandler {
	return &EditCompositionHandler{wines: w, varieties: v}
}

// Handle validates the command and persists the Wine's new Composition. It
// returns domain.ErrNotFound for an unknown Wine or Variety, and
// domain.ErrInvalidComposition when the proportions are empty or don't sum to
// ~100%.
func (h *EditCompositionHandler) Handle(ctx context.Context, cmd EditCompositionCommand) error {
	wine, err := h.wines.Get(ctx, cmd.WineID)
	if err != nil {
		return err
	}

	parts := make([]domain.CompositionPart, 0, len(cmd.Parts))
	for _, p := range cmd.Parts {
		if _, err := h.varieties.Get(ctx, p.VarietyID); err != nil {
			return err
		}
		parts = append(parts, domain.CompositionPart{VarietyID: p.VarietyID, Proportion: p.Proportion})
	}

	// Validate through the aggregate root: SetComposition enforces the invariant.
	// A Drinker editing the grapes by hand confirms them, so a future Style
	// re-seed must never clobber this Composition.
	if err := wine.SetComposition(parts, domain.ProvenanceConfirmed); err != nil {
		return err
	}

	return h.wines.SetComposition(ctx, wine.ID, wine.Composition)
}
