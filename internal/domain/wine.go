package domain

// Wine is a producer's product, identified by producer + name + Style and
// year-agnostic (the vintage lives on the Tasting). It is global reference data
// in the reference zone.
//
// A Wine has a Composition — the Varieties that make it up, with rough
// proportions. The Composition is a value object on this aggregate, and the
// Wine is its root: it is set and persisted through the Wine, never on its own.
// A Wine may exist before its Composition is known (seed-only data), so the
// zero Composition is allowed at construction; the invariant is enforced the
// moment a Composition is actually set, via SetComposition.
type Wine struct {
	ID          ID
	Producer    string
	Name        string
	Style       string
	Composition Composition
}

// NewWine constructs a Wine with a fresh ID and no Composition yet.
func NewWine(producer, name, style string) (Wine, error) {
	if producer == "" || name == "" {
		return Wine{}, ErrValidation
	}
	return Wine{ID: NewID(), Producer: producer, Name: name, Style: style}, nil
}

// SetComposition validates and sets the Wine's Composition, enforcing the
// aggregate invariant (≥1 Variety, proportions summing to ~100%). It returns
// ErrInvalidComposition and leaves the Wine unchanged on a bad Composition.
func (w *Wine) SetComposition(parts []CompositionPart) error {
	c, err := NewComposition(parts)
	if err != nil {
		return err
	}
	w.Composition = c
	return nil
}

// Label is the human-readable identity, e.g. "Penfolds — Bin 28 Shiraz".
func (w Wine) Label() string {
	return w.Producer + " — " + w.Name
}
