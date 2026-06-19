package app

import (
	"context"

	"go-wine/internal/domain"
)

// RecommendationView is the read-side view of one Discovery recommendation: an
// untried Variety the active Drinker is likely to enjoy, its resolved name, and
// the human-readable justification — the enjoyed grape(s) it sits nearest to
// ("Aglianico, because you loved Nebbiolo"). Not a domain entity.
type RecommendationView struct {
	VarietyID domain.ID
	Name      string
	// Because is the names of the enjoyed grape(s) that justify the
	// recommendation — the explainability the discovery design requires.
	Because []string
}

// DiscoveryHandler is the query-side service behind the Discovery page. It
// assembles the inputs the read-only domain Discovery service needs — the active
// Drinker's Taste profile (the enjoyed-grape set), the enjoyed grapes'
// Characteristics, and the untried candidate Varieties' Characteristics — calls
// domain.Discover, and maps the ranking to view-model DTOs with Variety names
// resolved for display. Every query is scoped to the active Drinker; the
// personal zone is never queried across owners.
//
// "Untried" means a Variety the Drinker has not reached through any wine they
// have tasted — i.e. not attributable through the Compositions of their drunk
// wines. The enjoyed grapes (a subset of the tried grapes) are therefore always
// excluded as candidates; so are minor blending grapes they merely encountered
// without enjoying, which is the conservative reading of "hasn't logged".
type DiscoveryHandler struct {
	wines     domain.WineRepository
	varieties domain.VarietyRepository
	tastings  domain.TastingRepository
}

func NewDiscoveryHandler(w domain.WineRepository, v domain.VarietyRepository, t domain.TastingRepository) *DiscoveryHandler {
	return &DiscoveryHandler{wines: w, varieties: v, tastings: t}
}

// Handle returns the active Drinker's recommendations, nearest first. An empty
// Taste profile yields no recommendations (the page surfaces the explanatory
// empty state). Varieties with no characteristics can't be placed in the space
// and are skipped by the domain service.
func (h *DiscoveryHandler) Handle(ctx context.Context, drinkerID domain.ID) ([]RecommendationView, error) {
	ts, err := h.tastings.ListByDrinker(ctx, drinkerID)
	if err != nil {
		return nil, err
	}

	// Compositions of the wines drunk: the inputs for both the Taste profile and
	// the tried-grape filter.
	comps := make(map[domain.ID]domain.Composition)
	tried := make(map[domain.ID]bool)
	for _, t := range ts {
		if _, ok := comps[t.WineID]; ok {
			continue
		}
		w, err := h.wines.Get(ctx, t.WineID)
		if err != nil {
			continue // a tasting of an unknown wine reaches no grapes
		}
		comps[t.WineID] = w.Composition
		for _, p := range w.Composition.Parts {
			tried[p.VarietyID] = true
		}
	}

	profile := domain.TasteProfile{}.Build(ts, comps)
	if len(profile.Enjoyed) == 0 {
		return nil, nil
	}

	// The enjoyed grapes' characteristics (the set we measure proximity to).
	enjoyedChars := make(map[domain.ID]domain.Characteristics, len(profile.Enjoyed))
	for _, e := range profile.Enjoyed {
		c, err := h.varieties.GetCharacteristics(ctx, e.VarietyID)
		if err != nil {
			return nil, err
		}
		enjoyedChars[e.VarietyID] = c
	}

	// The candidates: every Variety the Drinker has not tried, with its
	// characteristics. Names are kept for resolving the view.
	allVarieties, err := h.varieties.List(ctx)
	if err != nil {
		return nil, err
	}
	names := make(map[domain.ID]string, len(allVarieties))
	candidates := make(map[domain.ID]domain.Characteristics)
	for _, v := range allVarieties {
		names[v.ID] = v.Name
		if tried[v.ID] {
			continue
		}
		c, err := h.varieties.GetCharacteristics(ctx, v.ID)
		if err != nil {
			return nil, err
		}
		candidates[v.ID] = c
	}

	recs := domain.Discover(profile, enjoyedChars, candidates)
	out := make([]RecommendationView, 0, len(recs))
	for _, r := range recs {
		because := make([]string, 0, len(r.Because))
		for _, id := range r.Because {
			because = append(because, names[id])
		}
		out = append(out, RecommendationView{VarietyID: r.VarietyID, Name: names[r.VarietyID], Because: because})
	}
	return out, nil
}
