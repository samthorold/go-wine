package app

import (
	"context"
	"sort"

	"go-wine/internal/domain"
)

// VarietyView is a read-side view model for browsing Varieties: a grape's
// identity flattened for display. It is not a domain entity.
type VarietyView struct {
	ID   domain.ID
	Name string
}

// ListVarietiesHandler is a query-side service. It returns the global set of
// Varieties for the browse page, sorted by name.
type ListVarietiesHandler struct {
	varieties domain.VarietyRepository
}

func NewListVarietiesHandler(v domain.VarietyRepository) *ListVarietiesHandler {
	return &ListVarietiesHandler{varieties: v}
}

// Handle returns all Varieties as views, ordered by name.
func (h *ListVarietiesHandler) Handle(ctx context.Context) ([]VarietyView, error) {
	vs, err := h.varieties.List(ctx)
	if err != nil {
		return nil, err
	}
	views := make([]VarietyView, 0, len(vs))
	for _, v := range vs {
		views = append(views, VarietyView{ID: v.ID, Name: v.Name})
	}
	sort.Slice(views, func(i, j int) bool { return views[i].Name < views[j].Name })
	return views, nil
}
