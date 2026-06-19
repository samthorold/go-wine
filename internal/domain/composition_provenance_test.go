package domain

import "testing"

func mustComposition(t *testing.T, parts []CompositionPart, p Provenance) Composition {
	t.Helper()
	c, err := NewComposition(parts, p)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	return c
}

// When no Composition is stored yet, the seeded default applies wholesale.
func TestMergeComposition_AppliesSeedWhenAbsent(t *testing.T) {
	seeded := mustComposition(t, []CompositionPart{{VarietyID: NewID(), Proportion: 100}}, ProvenanceDefault)
	var stored Composition // zero/absent

	got := MergeComposition(seeded, stored)

	if len(got.Parts) != 1 || got.Provenance != ProvenanceDefault {
		t.Errorf("merge into absent = %+v, want the seeded default", got)
	}
}

// A re-seed over a still-default stored Composition updates it: the Style seed is
// the source of truth until the Drinker confirms the grapes themselves.
func TestMergeComposition_OverwritesStoredDefault(t *testing.T) {
	a, b := NewID(), NewID()
	stored := mustComposition(t, []CompositionPart{{VarietyID: a, Proportion: 100}}, ProvenanceDefault)
	seeded := mustComposition(t, []CompositionPart{{VarietyID: a, Proportion: 60}, {VarietyID: b, Proportion: 40}}, ProvenanceDefault)

	got := MergeComposition(seeded, stored)

	if len(got.Parts) != 2 {
		t.Errorf("a stored default should yield to the seed; got %+v", got)
	}
}

// THE no-clobber rule: a Drinker-confirmed Composition survives a re-seed
// untouched, the Style seed is ignored entirely.
func TestMergeComposition_NeverClobbersConfirmed(t *testing.T) {
	a, b := NewID(), NewID()
	stored := mustComposition(t, []CompositionPart{{VarietyID: a, Proportion: 100}}, ProvenanceConfirmed)
	seeded := mustComposition(t, []CompositionPart{{VarietyID: b, Proportion: 100}}, ProvenanceDefault)

	got := MergeComposition(seeded, stored)

	if len(got.Parts) != 1 || got.Parts[0].VarietyID != a || !got.IsConfirmed() {
		t.Errorf("confirmed Composition was clobbered: %+v", got)
	}
}
