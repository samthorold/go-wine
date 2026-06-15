package app

import (
	"context"
	"sort"
	"time"

	"go-wine/internal/domain"
)

// TastingView is a read-side view model: a Tasting flattened with the
// display fields a list page needs (e.g. the Wine label resolved). It is not a
// domain entity and deliberately does not round-trip back through the domain.
type TastingView struct {
	ID         domain.ID
	WineLabel  string
	Vintage    *int
	Rating     int
	Note       string
	Companions []string
	DrunkOn    time.Time
}

// ListTastingsHandler is a query-side service. It skips the aggregates and
// composes a view directly. (Against Postgres this becomes a single join; over
// the repository ports it resolves labels per row, which is fine at this scale.)
type ListTastingsHandler struct {
	wines      domain.WineRepository
	tastings   domain.TastingRepository
	companions domain.CompanionRepository
}

func NewListTastingsHandler(w domain.WineRepository, t domain.TastingRepository, c domain.CompanionRepository) *ListTastingsHandler {
	return &ListTastingsHandler{wines: w, tastings: t, companions: c}
}

// Handle returns a Drinker's Tastings, most recent first.
func (h *ListTastingsHandler) Handle(ctx context.Context, drinkerID domain.ID) ([]TastingView, error) {
	ts, err := h.tastings.ListByDrinker(ctx, drinkerID)
	if err != nil {
		return nil, err
	}

	// Resolve Companion names once for the Drinker's whole personal zone, scoped
	// to the active Drinker — never queried across owners.
	names := make(map[domain.ID]string)
	cs, err := h.companions.ListByDrinker(ctx, drinkerID)
	if err != nil {
		return nil, err
	}
	for _, c := range cs {
		names[c.ID] = c.Name
	}

	views := make([]TastingView, 0, len(ts))
	for _, t := range ts {
		label := "(unknown wine)"
		if w, err := h.wines.Get(ctx, t.WineID); err == nil {
			label = w.Label()
		}
		companions := make([]string, 0, len(t.Companions))
		for _, cid := range t.Companions {
			if n, ok := names[cid]; ok {
				companions = append(companions, n)
			}
		}
		views = append(views, TastingView{
			ID:         t.ID,
			WineLabel:  label,
			Vintage:    t.Vintage,
			Rating:     t.Rating.Int(),
			Note:       t.Note,
			Companions: companions,
			DrunkOn:    t.DrunkOn,
		})
	}

	sort.Slice(views, func(i, j int) bool {
		return views[i].DrunkOn.After(views[j].DrunkOn)
	})
	return views, nil
}
