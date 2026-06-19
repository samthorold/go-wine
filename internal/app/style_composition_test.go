package app_test

import (
	"context"
	"errors"
	"testing"

	"go-wine/internal/adapter/memory"
	"go-wine/internal/app"
	"go-wine/internal/domain"
	"go-wine/internal/seed"
)

// seedGrapes saves the named grapes into a fresh in-memory VarietyRepo and
// returns it plus a name→ID lookup, mirroring how each store mints its own
// random Variety IDs.
func seedGrapes(t *testing.T, names ...string) (*memory.VarietyRepo, map[string]domain.ID) {
	t.Helper()
	repo := memory.NewVarietyRepo()
	byName := make(map[string]domain.ID, len(names))
	for _, n := range names {
		v, err := domain.NewVariety(n)
		if err != nil {
			t.Fatalf("NewVariety(%q): %v", n, err)
		}
		repo.Save(v)
		byName[n] = v.ID
	}
	return repo, byName
}

// A known Style resolves to a default Composition over the seeded Varieties,
// joined by name, tagged ProvenanceDefault (a conventional guess).
func TestResolveStyleComposition_KnownStyle(t *testing.T) {
	repo, byName := seedGrapes(t, "Sangiovese", "Merlot")
	h := app.NewResolveStyleCompositionHandler(repo, seed.StyleCompositions())

	c, err := h.Handle(context.Background(), "Chianti")
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if c.Provenance != domain.ProvenanceDefault {
		t.Errorf("Provenance = %v, want default", c.Provenance)
	}
	if len(c.Parts) == 0 {
		t.Fatal("expected a non-empty default Composition for Chianti")
	}
	// Chianti is Sangiovese-dominant: the lead grape must be Sangiovese.
	lead := c.Parts[0]
	for _, p := range c.Parts {
		if p.Proportion > lead.Proportion {
			lead = p
		}
	}
	if lead.VarietyID != byName["Sangiovese"] {
		t.Errorf("lead grape = %v, want Sangiovese", lead.VarietyID)
	}
}

// An unknown Style has no conventional default — ErrNotFound, not a silent empty
// Composition.
func TestResolveStyleComposition_UnknownStyle(t *testing.T) {
	repo, _ := seedGrapes(t, "Sangiovese")
	h := app.NewResolveStyleCompositionHandler(repo, seed.StyleCompositions())

	if _, err := h.Handle(context.Background(), "Nonexistent Style"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

// The map is keyed by Variety name; a store missing a grape the map references
// simply omits that grape rather than failing — but the remaining grapes must
// still form a valid Composition (sum ~100%). Here Chablis is 100% Chardonnay.
func TestResolveStyleComposition_JoinsByName(t *testing.T) {
	repo, byName := seedGrapes(t, "Chardonnay")
	h := app.NewResolveStyleCompositionHandler(repo, seed.StyleCompositions())

	c, err := h.Handle(context.Background(), "Chablis")
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if len(c.Parts) != 1 || c.Parts[0].VarietyID != byName["Chardonnay"] || c.Parts[0].Proportion != 100 {
		t.Errorf("Chablis = %+v, want 100%% Chardonnay", c.Parts)
	}
}
