package app

import (
	"context"

	"go-wine/internal/domain"
)

// StylePartSeed is one grape's conventional share of a Style's default
// Composition, keyed by Variety name (the stable join key, since each store mints
// its own Variety IDs).
type StylePartSeed struct {
	Variety    string
	Proportion int
}

// StyleSeed is one Style's default Composition: the conventional blend a label's
// Style implies, over Varieties named by the seed. It carries no Provenance — the
// seed always proposes a default; MergeComposition decides whether it lands.
type StyleSeed struct {
	Style string
	Parts []StylePartSeed
}

// ResolveStyleCompositionHandler is a query-side service over the Style →
// default Composition seed. Given a Style string it resolves the conventional
// blend to a domain.Composition (Provenance=default) over the Varieties the store
// actually carries, joining the seed's grape names to stored Variety IDs.
//
// This is read-only reference resolution — it never writes — so it skips the Wine
// aggregate and works straight off the VarietyRepository, exactly like the other
// query-side services.
type ResolveStyleCompositionHandler struct {
	varieties domain.VarietyRepository
	seed      []StyleSeed
}

func NewResolveStyleCompositionHandler(v domain.VarietyRepository, seed []StyleSeed) *ResolveStyleCompositionHandler {
	return &ResolveStyleCompositionHandler{varieties: v, seed: seed}
}

// Handle resolves a Style to its default Composition. An unrecognised Style — one
// with no conventional default in the seed — is domain.ErrNotFound, so the caller
// can distinguish "no default known" from a valid empty result. A grape the seed
// names but the store doesn't carry is dropped; the resulting Composition is
// validated through the domain (so a join that leaves the proportions off-balance
// surfaces as ErrInvalidComposition rather than a silent bad default).
func (h *ResolveStyleCompositionHandler) Handle(ctx context.Context, style string) (domain.Composition, error) {
	var seed *StyleSeed
	for i := range h.seed {
		if h.seed[i].Style == style {
			seed = &h.seed[i]
			break
		}
	}
	if seed == nil {
		return domain.Composition{}, domain.ErrNotFound
	}

	vs, err := h.varieties.List(ctx)
	if err != nil {
		return domain.Composition{}, err
	}
	idByName := make(map[string]domain.ID, len(vs))
	for _, v := range vs {
		idByName[v.Name] = v.ID
	}

	parts := make([]domain.CompositionPart, 0, len(seed.Parts))
	for _, p := range seed.Parts {
		id, ok := idByName[p.Variety]
		if !ok {
			continue // the store doesn't carry this grape; drop it from the default
		}
		parts = append(parts, domain.CompositionPart{VarietyID: id, Proportion: p.Proportion})
	}

	return domain.NewComposition(parts, domain.ProvenanceDefault)
}
