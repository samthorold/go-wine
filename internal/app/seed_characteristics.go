package app

import (
	"context"

	"go-wine/internal/domain"
)

// VarietySeed is one grape's default characteristics from the seed rubric: the
// five scalar axes on the 1..5 scale plus the typical flavour-note tags, keyed
// by Variety name. The rubric is keyed by name rather than ID because each store
// mints its own random Variety IDs when it seeds the grape names; the name is
// the stable join. It carries no Provenance — the seed always proposes a
// default; the merge decides whether it lands.
type VarietySeed struct {
	Name                                      string
	Body, Tannin, Acidity, Sweetness, Alcohol int
	Notes                                     []string
}

// SeedCharacteristics runs the provenance seed-merge over a rubric against any
// VarietyRepository — used by both the Postgres and in-memory stores so the same
// no-clobber rule holds on each. It resolves each rubric entry to a stored
// Variety by name, builds the seeded default bundle, reads whatever is currently
// stored, and persists domain.MergeCharacteristics of the two: confirmed values
// are preserved untouched, while absent or still-default values take the seed.
// Re-running it is therefore idempotent and non-clobbering.
//
// A grape whose stored value is already confirmed is left entirely alone (no
// write at all), so a re-seed touches only what it may. A rubric entry naming a
// grape the store doesn't have is skipped — the seed only describes grapes that
// exist.
func SeedCharacteristics(ctx context.Context, repo domain.VarietyRepository, rubric []VarietySeed) error {
	vs, err := repo.List(ctx)
	if err != nil {
		return err
	}
	idByName := make(map[string]domain.ID, len(vs))
	for _, v := range vs {
		idByName[v.Name] = v.ID
	}

	for _, s := range rubric {
		id, ok := idByName[s.Name]
		if !ok {
			continue // the seed only describes grapes the store actually has
		}
		seeded, err := domain.NewCharacteristics(s.Body, s.Tannin, s.Acidity, s.Sweetness, s.Alcohol, s.Notes, domain.ProvenanceDefault)
		if err != nil {
			return err
		}
		stored, err := repo.GetCharacteristics(ctx, id)
		if err != nil {
			return err
		}
		if stored.IsConfirmed() {
			continue // never clobber a vetted value; leave it untouched
		}
		merged := domain.MergeCharacteristics(seeded, stored)
		if err := repo.SetCharacteristics(ctx, id, merged); err != nil {
			return err
		}
	}
	return nil
}
