package domain

// Wine is a producer's product, identified by producer + name + Style and
// year-agnostic (the vintage lives on the Tasting). It is global reference data
// in the reference zone.
//
// Composition (the Varieties that make up the Wine) is a later vertical slice;
// this first slice carries only what logging a Tasting needs to reference and
// display a Wine.
type Wine struct {
	ID       ID
	Producer string
	Name     string
	Style    string
}

// NewWine constructs a Wine with a fresh ID.
func NewWine(producer, name, style string) (Wine, error) {
	if producer == "" || name == "" {
		return Wine{}, ErrValidation
	}
	return Wine{ID: NewID(), Producer: producer, Name: name, Style: style}, nil
}

// Label is the human-readable identity, e.g. "Penfolds — Bin 28 Shiraz".
func (w Wine) Label() string {
	return w.Producer + " — " + w.Name
}
