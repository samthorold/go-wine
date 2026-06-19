package domain

// Variety is a grape variety — Shiraz, Pinot Noir, Riesling. The portable axis
// along which taste is learned and discovery happens. It is global reference
// data in the reference zone.
//
// The Variety struct carries only its identity (a name) and stays a plain
// comparable value. Its intrinsic Characteristics — the scalar axes, the
// flavour-note tags, and their Provenance — are a value object owned by the
// VarietyRepository alongside it (one repository per aggregate root), kept off
// this struct so it remains ==-comparable. See characteristics.go.
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
