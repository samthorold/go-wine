package app

import (
	"context"
	"errors"

	"go-wine/internal/domain"
)

// SeedStyleCompositions runs the provenance seed-merge over the Style → default
// Composition map against any Wine/Variety repositories — used by both the
// Postgres and in-memory stores so the same no-clobber rule holds on each. The
// sibling of SeedCharacteristics on the output side.
//
// For every Wine whose Style has a conventional default, it resolves that default
// (joined to stored Varieties by name), reads whatever Composition is currently
// stored, and persists domain.MergeComposition of the two: a Drinker-confirmed
// Composition is preserved untouched, while an absent or still-default one takes
// the Style seed. Re-running it is therefore idempotent and non-clobbering, so a
// corrected map flows through to unconfirmed Wines while confirmed grapes survive.
//
// A Wine whose Style has no entry in the map, or whose resolved default would be
// invalid (e.g. the store carries none of the Style's grapes), is left entirely
// alone — the seed only fills a default it can actually stand behind.
func SeedStyleCompositions(ctx context.Context, wines domain.WineRepository, varieties domain.VarietyRepository, styleSeed []StyleSeed) error {
	resolver := NewResolveStyleCompositionHandler(varieties, styleSeed)

	ws, err := wines.List(ctx)
	if err != nil {
		return err
	}
	for _, w := range ws {
		seeded, err := resolver.Handle(ctx, w.Style)
		if errors.Is(err, domain.ErrNotFound) || errors.Is(err, domain.ErrInvalidComposition) {
			continue // no conventional default this store can stand behind
		}
		if err != nil {
			return err
		}
		if w.Composition.IsConfirmed() {
			continue // never clobber a vetted Composition; leave it untouched
		}
		merged := domain.MergeComposition(seeded, w.Composition)
		if err := wines.SetComposition(ctx, w.ID, merged); err != nil {
			return err
		}
	}
	return nil
}
