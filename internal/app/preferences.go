package app

import (
	"context"
	"sort"

	"go-wine/internal/domain"
)

// WineVerdictView is the read-side view of "do I like this Wine?": the active
// Drinker's aggregate rating of a Wine. Tasted is false when they have not
// tasted it yet — a clear unknown rather than a fake zero. Not a domain entity.
type WineVerdictView struct {
	Tasted     bool
	Count      int
	MeanRating float64
}

// VarietyPreferenceView is the read-side view of "do I like this Variety?": the
// active Drinker's preference derived through the Compositions of the wines they
// have drunk. Tasted is false when no drunk wine contains the Variety.
type VarietyPreferenceView struct {
	Tasted     bool
	Preference float64
}

// EnjoyedVarietyView is one grape in the Taste profile, with the Variety's name
// resolved for display and its enjoyment Weight.
type EnjoyedVarietyView struct {
	VarietyID domain.ID
	Name      string
	Weight    float64
}

// TasteProfileView is the read-side view of the active Drinker's Taste profile:
// the SET of enjoyed grapes, each weighted by enjoyment, ordered by weight then
// name. It is deliberately a collection — a region of characteristics space,
// possibly several clusters — not a single averaged point, which is the property
// Discovery's proximity-to-a-set search depends on.
type TasteProfileView struct {
	Enjoyed []EnjoyedVarietyView
}

// PreferencesHandler is a query-side service computing the do-I-like read models
// and the Taste profile. It fetches via the repository ports, hands the data to
// the read-only domain services (no aggregate mutation), and returns view-model
// DTOs. Every query is scoped to a Drinker — the personal zone is never queried
// across owners.
type PreferencesHandler struct {
	wines     domain.WineRepository
	varieties domain.VarietyRepository
	tastings  domain.TastingRepository
}

func NewPreferencesHandler(w domain.WineRepository, v domain.VarietyRepository, t domain.TastingRepository) *PreferencesHandler {
	return &PreferencesHandler{wines: w, varieties: v, tastings: t}
}

// WineVerdict aggregates the active Drinker's Tastings of one Wine.
func (h *PreferencesHandler) WineVerdict(ctx context.Context, drinkerID, wineID domain.ID) (WineVerdictView, error) {
	ts, err := h.tastings.ListByDrinker(ctx, drinkerID)
	if err != nil {
		return WineVerdictView{}, err
	}
	ofWine := make([]domain.Tasting, 0, len(ts))
	for _, t := range ts {
		if t.WineID == wineID {
			ofWine = append(ofWine, t)
		}
	}
	v := domain.LikeWine(ofWine)
	return WineVerdictView{Tasted: v.Tasted, Count: v.Count, MeanRating: v.MeanRating}, nil
}

// VarietyPreference derives the active Drinker's preference for a Variety,
// attributed through the Compositions of the wines they have drunk.
func (h *PreferencesHandler) VarietyPreference(ctx context.Context, drinkerID, varietyID domain.ID) (VarietyPreferenceView, error) {
	ts, comps, err := h.tastingsAndCompositions(ctx, drinkerID)
	if err != nil {
		return VarietyPreferenceView{}, err
	}
	p := domain.LikeVariety(varietyID, ts, comps)
	return VarietyPreferenceView{Tasted: p.Tasted, Preference: p.Preference}, nil
}

// TasteProfile builds the active Drinker's enjoyed-grape set, resolving Variety
// names for display and ordering by descending weight then name.
func (h *PreferencesHandler) TasteProfile(ctx context.Context, drinkerID domain.ID) (TasteProfileView, error) {
	ts, comps, err := h.tastingsAndCompositions(ctx, drinkerID)
	if err != nil {
		return TasteProfileView{}, err
	}
	profile := domain.TasteProfile{}.Build(ts, comps)
	out := make([]EnjoyedVarietyView, 0, len(profile.Enjoyed))
	for _, e := range profile.Enjoyed {
		name := ""
		if v, err := h.varieties.Get(ctx, e.VarietyID); err == nil {
			name = v.Name
		}
		out = append(out, EnjoyedVarietyView{VarietyID: e.VarietyID, Name: name, Weight: e.Weight})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Weight != out[j].Weight {
			return out[i].Weight > out[j].Weight
		}
		return out[i].Name < out[j].Name
	})
	return TasteProfileView{Enjoyed: out}, nil
}

// tastingsAndCompositions fetches the active Drinker's Tastings and the
// Compositions of every Wine those Tastings reference, the inputs the
// attribution read services need.
func (h *PreferencesHandler) tastingsAndCompositions(ctx context.Context, drinkerID domain.ID) ([]domain.Tasting, map[domain.ID]domain.Composition, error) {
	ts, err := h.tastings.ListByDrinker(ctx, drinkerID)
	if err != nil {
		return nil, nil, err
	}
	comps := make(map[domain.ID]domain.Composition)
	for _, t := range ts {
		if _, ok := comps[t.WineID]; ok {
			continue
		}
		w, err := h.wines.Get(ctx, t.WineID)
		if err != nil {
			continue // a tasting of an unknown wine contributes no grapes
		}
		comps[t.WineID] = w.Composition
	}
	return ts, comps, nil
}
