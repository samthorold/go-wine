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
type Composition struct {
	Parts []CompositionPart
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
func NewComposition(parts []CompositionPart) (Composition, error) {
	if len(parts) == 0 {
		return Composition{}, ErrInvalidComposition
	}
	sum := 0
	for _, p := range parts {
		if p.VarietyID == "" || p.Proportion < 1 || p.Proportion > 100 {
			return Composition{}, ErrInvalidComposition
		}
		sum += p.Proportion
	}
	if sum < 100-compositionTolerance || sum > 100+compositionTolerance {
		return Composition{}, ErrInvalidComposition
	}
	return Composition{Parts: parts}, nil
}

// IsZero reports whether a Wine has no Composition set yet.
func (c Composition) IsZero() bool { return len(c.Parts) == 0 }
