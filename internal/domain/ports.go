package domain

import "context"

// Repository ports. One per aggregate root, not per table. Implemented by the
// in-memory adapter (fast unit tests) and the Postgres adapter (integration).

// DrinkerRepository owns Drinkers.
type DrinkerRepository interface {
	Get(ctx context.Context, id ID) (Drinker, error)
	List(ctx context.Context) ([]Drinker, error)
}

// WineRepository owns Wines (and, in later slices, their Composition).
type WineRepository interface {
	Get(ctx context.Context, id ID) (Wine, error)
	List(ctx context.Context) ([]Wine, error)
}

// TastingRepository owns Tastings. Reads are always scoped to a Drinker — the
// personal zone is never queried across owners.
type TastingRepository interface {
	Add(ctx context.Context, t Tasting) error
	ListByDrinker(ctx context.Context, drinkerID ID) ([]Tasting, error)
}
