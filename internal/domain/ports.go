package domain

import "context"

// Repository ports. One per aggregate root, not per table. Implemented by the
// in-memory adapter (fast unit tests) and the Postgres adapter (integration).

// DrinkerRepository owns Drinkers. Save upserts: it inserts a new Drinker or
// updates an existing one (a rename), keyed by ID.
type DrinkerRepository interface {
	Save(ctx context.Context, d Drinker) error
	Get(ctx context.Context, id ID) (Drinker, error)
	List(ctx context.Context) ([]Drinker, error)
}

// WineRepository owns the Wine aggregate — the Wine and its Composition,
// persisted together. SetComposition replaces the Wine's Composition (the
// caller has already validated it through the domain); Get and List return the
// Wine with its Composition attached.
type WineRepository interface {
	Get(ctx context.Context, id ID) (Wine, error)
	List(ctx context.Context) ([]Wine, error)
	SetComposition(ctx context.Context, wineID ID, c Composition) error
}

// VarietyRepository owns the Variety aggregate — the grape's identity and its
// intrinsic Characteristics, persisted together. Global reference data, not
// scoped to a Drinker. GetCharacteristics returns the zero bundle (IsZero) for a
// Variety not yet seeded; SetCharacteristics replaces the stored bundle and
// returns ErrNotFound for an unknown Variety.
type VarietyRepository interface {
	Get(ctx context.Context, id ID) (Variety, error)
	List(ctx context.Context) ([]Variety, error)
	GetCharacteristics(ctx context.Context, id ID) (Characteristics, error)
	SetCharacteristics(ctx context.Context, id ID, c Characteristics) error
}

// TastingRepository owns Tastings. Reads are always scoped to a Drinker — the
// personal zone is never queried across owners.
type TastingRepository interface {
	Add(ctx context.Context, t Tasting) error
	ListByDrinker(ctx context.Context, drinkerID ID) ([]Tasting, error)
}

// CompanionRepository owns Companions — personal-zone reference data. Like
// Tastings, Companions are always scoped to a Drinker; they are never queried
// across owners.
type CompanionRepository interface {
	Add(ctx context.Context, c Companion) error
	Get(ctx context.Context, id ID) (Companion, error)
	ListByDrinker(ctx context.Context, drinkerID ID) ([]Companion, error)
}
