package domain

import "testing"

func mustChars(t *testing.T, body, tannin, acidity, sweetness, alcohol int, notes []string, p Provenance) Characteristics {
	t.Helper()
	c, err := NewCharacteristics(body, tannin, acidity, sweetness, alcohol, notes, p)
	if err != nil {
		t.Fatalf("NewCharacteristics: %v", err)
	}
	return c
}

// When nothing is stored yet, the seeded default applies wholesale.
func TestMergeCharacteristics_AppliesSeedWhenAbsent(t *testing.T) {
	seeded := mustChars(t, 5, 4, 2, 1, 5, []string{"blackberry"}, ProvenanceDefault)
	var stored Characteristics // zero/absent

	got := MergeCharacteristics(seeded, stored)

	if got.Body != seeded.Body || len(got.Notes) != 1 || got.Notes[0] != "blackberry" {
		t.Errorf("merge into absent = %+v, want the seeded default", got)
	}
	if got.Provenance != ProvenanceDefault {
		t.Errorf("Provenance = %v, want default", got.Provenance)
	}
}

// A re-seed over a still-default stored value updates it: the seed is the source
// of truth until the Drinker confirms. This keeps re-seeding idempotent and lets
// a corrected rubric flow through to unconfirmed grapes.
func TestMergeCharacteristics_OverwritesStoredDefault(t *testing.T) {
	stored := mustChars(t, 1, 1, 1, 1, 1, []string{"old"}, ProvenanceDefault)
	seeded := mustChars(t, 5, 4, 2, 1, 5, []string{"new"}, ProvenanceDefault)

	got := MergeCharacteristics(seeded, stored)

	if got.Body.Int() != 5 || got.Tannin.Int() != 4 {
		t.Errorf("a stored default should yield to the seed; got %+v", got)
	}
	if got.Provenance != ProvenanceDefault {
		t.Errorf("Provenance = %v, want default", got.Provenance)
	}
}

// THE no-clobber rule: a confirmed stored value is preserved untouched, the seed
// is ignored entirely. This is the genuine command-side domain rule.
func TestMergeCharacteristics_NeverClobbersConfirmed(t *testing.T) {
	stored := mustChars(t, 2, 3, 4, 1, 2, []string{"mine"}, ProvenanceConfirmed)
	seeded := mustChars(t, 5, 5, 5, 5, 5, []string{"seed"}, ProvenanceDefault)

	got := MergeCharacteristics(seeded, stored)

	if got.Body.Int() != 2 || got.Tannin.Int() != 3 || got.Acidity.Int() != 4 {
		t.Errorf("confirmed values were clobbered: %+v", got)
	}
	if len(got.Notes) != 1 || got.Notes[0] != "mine" {
		t.Errorf("confirmed notes were clobbered: %v", got.Notes)
	}
	if got.Provenance != ProvenanceConfirmed {
		t.Errorf("Provenance = %v, want confirmed (still vetted)", got.Provenance)
	}
}
