package domain

import "time"

// Tasting is a single logged event of drinking a Wine — the central thing a
// Drinker creates. The Rating lives here and only here; "do I like this Wine?"
// is always computed by aggregating Tastings. Context (food, mood, weather,
// anecdote) is captured in the free-text Note; everything else about the
// occasion that stays explicit is here.
type Tasting struct {
	ID        ID
	DrinkerID ID
	WineID    ID
	Vintage   *int // year, optional
	Rating    Rating
	Note      string
	Companions []ID
	DrunkOn   time.Time
}

// NewTasting constructs a Tasting with a fresh ID. Callers pass an
// already-validated Rating value object, so construction cannot produce an
// out-of-range rating.
func NewTasting(drinkerID, wineID ID, vintage *int, rating Rating, note string, companions []ID, drunkOn time.Time) (Tasting, error) {
	if drinkerID == "" || wineID == "" {
		return Tasting{}, ErrValidation
	}
	if drunkOn.IsZero() {
		return Tasting{}, ErrValidation
	}
	return Tasting{
		ID:         NewID(),
		DrinkerID:  drinkerID,
		WineID:     wineID,
		Vintage:    vintage,
		Rating:     rating,
		Note:       note,
		Companions: companions,
		DrunkOn:    drunkOn,
	}, nil
}
