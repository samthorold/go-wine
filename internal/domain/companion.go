package domain

// Companion is a named person a Drinker was with for a Tasting. It is
// personal-zone reference data — just a name — scoped to the Drinker who owns
// it and attachable to many of that Drinker's Tastings. A Companion is *never*
// a Drinker: even if the same person also uses the app, to its owner it stays a
// name attached to Tastings, which is what keeps the app free of any
// cross-Drinker sharing or consent.
type Companion struct {
	ID        ID
	DrinkerID ID
	Name      string
}

// NewCompanion constructs a Companion with a fresh ID, scoped to a Drinker.
func NewCompanion(drinkerID ID, name string) (Companion, error) {
	if drinkerID == "" || name == "" {
		return Companion{}, ErrValidation
	}
	return Companion{ID: NewID(), DrinkerID: drinkerID, Name: name}, nil
}
