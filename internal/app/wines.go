package app

import (
	"context"
	"sort"

	"go-wine/internal/domain"
)

// WineView is a read-side view model for browsing Wines: a Wine's identity
// flattened for display, plus whether it has a Composition set yet. Not a
// domain entity.
type WineView struct {
	ID             domain.ID
	Label          string
	Style          string
	HasComposition bool
}

// CompositionPartView is one Variety's share of a Wine, with the Variety's name
// resolved for display.
type CompositionPartView struct {
	VarietyID   domain.ID
	VarietyName string
	Proportion  int
}

// WineDetailView is the read-side view for a single Wine, including its
// Composition with Variety names resolved.
type WineDetailView struct {
	ID          domain.ID
	Producer    string
	Name        string
	Label       string
	Style       string
	Composition []CompositionPartView
}

// ListWinesHandler is a query-side service: it returns all Wines for the browse
// page, sorted by label.
type ListWinesHandler struct {
	wines domain.WineRepository
}

func NewListWinesHandler(w domain.WineRepository) *ListWinesHandler {
	return &ListWinesHandler{wines: w}
}

// Handle returns all Wines as views, ordered by label.
func (h *ListWinesHandler) Handle(ctx context.Context) ([]WineView, error) {
	ws, err := h.wines.List(ctx)
	if err != nil {
		return nil, err
	}
	views := make([]WineView, 0, len(ws))
	for _, w := range ws {
		views = append(views, WineView{
			ID:             w.ID,
			Label:          w.Label(),
			Style:          w.Style,
			HasComposition: !w.Composition.IsZero(),
		})
	}
	sort.Slice(views, func(i, j int) bool { return views[i].Label < views[j].Label })
	return views, nil
}

// GetWineHandler is a query-side service: it returns one Wine with its
// Composition, resolving each part's Variety name for display.
type GetWineHandler struct {
	wines     domain.WineRepository
	varieties domain.VarietyRepository
}

func NewGetWineHandler(w domain.WineRepository, v domain.VarietyRepository) *GetWineHandler {
	return &GetWineHandler{wines: w, varieties: v}
}

// Handle returns the Wine detail view, or domain.ErrNotFound if the Wine does
// not exist. Composition parts are ordered by descending share.
func (h *GetWineHandler) Handle(ctx context.Context, id domain.ID) (WineDetailView, error) {
	w, err := h.wines.Get(ctx, id)
	if err != nil {
		return WineDetailView{}, err
	}
	parts := make([]CompositionPartView, 0, len(w.Composition.Parts))
	for _, p := range w.Composition.Parts {
		name := ""
		if v, err := h.varieties.Get(ctx, p.VarietyID); err == nil {
			name = v.Name
		}
		parts = append(parts, CompositionPartView{VarietyID: p.VarietyID, VarietyName: name, Proportion: p.Proportion})
	}
	sort.Slice(parts, func(i, j int) bool { return parts[i].Proportion > parts[j].Proportion })
	return WineDetailView{
		ID:          w.ID,
		Producer:    w.Producer,
		Name:        w.Name,
		Label:       w.Label(),
		Style:       w.Style,
		Composition: parts,
	}, nil
}
