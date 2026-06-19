package domain

// Composition is the set of Varieties that make up a Wine, with rough
// proportions. It is a value object on the Wine aggregate, not an entity of its
// own: it has no identity and is only ever meaningful as part of a Wine. The
// command-side invariant — at least one Variety, proportions summing to ~100% —
// lives here, making this the first genuine aggregate rule in the domain.
//
// Proportions are integer percentages (1..100). "Rough" is the operative word
// in the ubiquitous language ("mostly Shiraz"), so an exact sum to 100 is not
// demanded: the sum must land within compositionTolerance of 100. That slack
// lets equal-thirds blends (33/33/33 = 99) and rounded estimates through while
// still rejecting a Composition that plainly doesn't add up.
// A Composition carries a single binary Provenance, exactly as Characteristics
// does: 'default' when filled from the Style → Composition seed (a conventional
// guess), 'confirmed' once the Drinker sets or edits the grapes themselves. The
// seed-merge (MergeComposition) preserves a confirmed Composition across
// re-seeds; a default one tracks the current Style seed.
type Composition struct {
	Parts      []CompositionPart
	Provenance Provenance
}

// CompositionPart is one Variety's share of a Wine, as an integer percentage.
type CompositionPart struct {
	VarietyID  ID
	Proportion int
}

// compositionTolerance is how far the proportion sum may stray from 100 and
// still be accepted, in percentage points either side.
const compositionTolerance = 2

// NewComposition constructs a validated Composition. It rejects an empty
// Composition (severing the Wine→Composition→Variety chain that Discovery walks)
// and any proportion sum outside the tolerance band around 100%.
func NewComposition(parts []CompositionPart, p Provenance) (Composition, error) {
	if len(parts) == 0 {
		return Composition{}, ErrInvalidComposition
	}
	sum := 0
	for _, part := range parts {
		if part.VarietyID == "" || part.Proportion < 1 || part.Proportion > 100 {
			return Composition{}, ErrInvalidComposition
		}
		sum += part.Proportion
	}
	if sum < 100-compositionTolerance || sum > 100+compositionTolerance {
		return Composition{}, ErrInvalidComposition
	}
	return Composition{Parts: parts, Provenance: p}, nil
}

// IsZero reports whether a Wine has no Composition set yet.
func (c Composition) IsZero() bool { return len(c.Parts) == 0 }

// IsConfirmed reports whether the Drinker has set or vetted the grapes
// themselves, so a re-seed must never clobber them.
func (c Composition) IsConfirmed() bool { return c.Provenance == ProvenanceConfirmed }
