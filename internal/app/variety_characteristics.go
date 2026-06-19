package app

import (
	"context"

	"go-wine/internal/domain"
)

// VarietyDetailView is the read-side view for a single Variety: its name plus
// its intrinsic Characteristics flattened for display. Confirmed reports whether
// the Drinker has vetted the bundle (Provenance confirmed). Not a domain entity.
type VarietyDetailView struct {
	ID                                        domain.ID
	Name                                      string
	Body, Tannin, Acidity, Sweetness, Alcohol int
	Notes                                     []string
	Confirmed                                 bool
	// HasCharacteristics is false for a Variety not yet seeded — the axes are off
	// the rubric and should not be shown as real values.
	HasCharacteristics bool
}

// GetVarietyHandler is a query-side service: it returns one Variety with its
// Characteristics, for the detail page and the edit form.
type GetVarietyHandler struct {
	varieties domain.VarietyRepository
}

func NewGetVarietyHandler(v domain.VarietyRepository) *GetVarietyHandler {
	return &GetVarietyHandler{varieties: v}
}

// Handle returns the Variety detail view, or domain.ErrNotFound if the Variety
// does not exist.
func (h *GetVarietyHandler) Handle(ctx context.Context, id domain.ID) (VarietyDetailView, error) {
	v, err := h.varieties.Get(ctx, id)
	if err != nil {
		return VarietyDetailView{}, err
	}
	c, err := h.varieties.GetCharacteristics(ctx, id)
	if err != nil {
		return VarietyDetailView{}, err
	}
	view := VarietyDetailView{ID: v.ID, Name: v.Name}
	if !c.IsZero() {
		view.HasCharacteristics = true
		view.Body = c.Body.Int()
		view.Tannin = c.Tannin.Int()
		view.Acidity = c.Acidity.Int()
		view.Sweetness = c.Sweetness.Int()
		view.Alcohol = c.Alcohol.Int()
		view.Notes = c.Notes
		view.Confirmed = c.IsConfirmed()
	}
	return view, nil
}

// EditCharacteristicsCommand is the input to hand-editing a Variety's
// Characteristics. The axes are plain ints; the domain validates them against
// the 1..5 rubric. Notes is the flavour-note tag set.
type EditCharacteristicsCommand struct {
	VarietyID                                 domain.ID
	Body, Tannin, Acidity, Sweetness, Alcohol int
	Notes                                     []string
}

// EditCharacteristicsHandler is a command-side use case. A hand-edit is the rare
// override the design allows: it goes through the domain (the axes are validated
// against the rubric) and marks the bundle Provenance confirmed, which is what
// protects it from being clobbered by a future re-seed (see
// domain.MergeCharacteristics).
type EditCharacteristicsHandler struct {
	varieties domain.VarietyRepository
}

func NewEditCharacteristicsHandler(v domain.VarietyRepository) *EditCharacteristicsHandler {
	return &EditCharacteristicsHandler{varieties: v}
}

// Handle validates the command and persists the Variety's Characteristics,
// marked confirmed. It returns domain.ErrNotFound for an unknown Variety and
// domain.ErrInvalidCharacteristics when an axis is off the 1..5 rubric.
func (h *EditCharacteristicsHandler) Handle(ctx context.Context, cmd EditCharacteristicsCommand) error {
	if _, err := h.varieties.Get(ctx, cmd.VarietyID); err != nil {
		return err
	}
	c, err := domain.NewCharacteristics(cmd.Body, cmd.Tannin, cmd.Acidity, cmd.Sweetness, cmd.Alcohol, cmd.Notes, domain.ProvenanceConfirmed)
	if err != nil {
		return err
	}
	return h.varieties.SetCharacteristics(ctx, cmd.VarietyID, c)
}
