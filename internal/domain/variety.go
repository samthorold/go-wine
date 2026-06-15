package domain

// Variety is a grape variety — Shiraz, Pinot Noir, Riesling. The portable axis
// along which taste is learned and discovery happens. It is global reference
// data in the reference zone. This first slice carries only its identity (a
// name); characteristics and provenance arrive in a later slice.
type Variety struct {
	ID   ID
	Name string
}

// NewVariety constructs a Variety with a fresh ID.
func NewVariety(name string) (Variety, error) {
	if name == "" {
		return Variety{}, ErrValidation
	}
	return Variety{ID: NewID(), Name: name}, nil
}
