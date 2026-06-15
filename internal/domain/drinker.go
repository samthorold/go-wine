package domain

// Drinker owns a personal zone — Tastings, ratings, Companions, Taste profile.
// Several exist from the start; the app selects an active one through a plain
// switcher rather than authentication. A Drinker is an identity for ownership,
// not a secured account.
type Drinker struct {
	ID   ID
	Name string
}

// NewDrinker constructs a Drinker with a fresh ID.
func NewDrinker(name string) (Drinker, error) {
	if name == "" {
		return Drinker{}, ErrValidation
	}
	return Drinker{ID: NewID(), Name: name}, nil
}
