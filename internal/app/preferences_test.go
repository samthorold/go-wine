package app_test

import (
	"context"
	"testing"
	"time"

	"go-wine/internal/adapter/memory"
	"go-wine/internal/app"
	"go-wine/internal/domain"
)

// seedTasting adds a Tasting of a Wine for a Drinker at the given rating.
func seedTasting(t *testing.T, tastings *memory.TastingRepo, drinkerID, wineID domain.ID, rating int) {
	t.Helper()
	r, _ := domain.NewRating(rating)
	ta, err := domain.NewTasting(drinkerID, wineID, nil, r, "", nil, time.Now())
	if err != nil {
		t.Fatalf("NewTasting: %v", err)
	}
	if err := tastings.Add(context.Background(), ta); err != nil {
		t.Fatalf("Add: %v", err)
	}
}

// The Wine verdict handler aggregates the active Drinker's Tastings of a Wine,
// scoped to that Drinker.
func TestPreferences_WineVerdictAggregatesDrinkersTastings(t *testing.T) {
	wines := memory.NewWineRepo()
	tastings := memory.NewTastingRepo()
	varieties := memory.NewVarietyRepo()

	w, _ := domain.NewWine("Penfolds", "Bin 28 Shiraz", "Shiraz")
	wines.Save(w)
	sam, _ := domain.NewDrinker("Sam")
	other, _ := domain.NewDrinker("Other")

	seedTasting(t, tastings, sam.ID, w.ID, 4)
	seedTasting(t, tastings, sam.ID, w.ID, 5)
	// Another Drinker's tasting of the same Wine must not leak in.
	seedTasting(t, tastings, other.ID, w.ID, 1)

	h := app.NewPreferencesHandler(wines, varieties, tastings)
	got, err := h.WineVerdict(context.Background(), sam.ID, w.ID)
	if err != nil {
		t.Fatalf("WineVerdict: %v", err)
	}
	if !got.Tasted || got.Count != 2 || got.MeanRating != 4.5 {
		t.Errorf("verdict = %+v, want tasted, count 2, mean 4.5", got)
	}
}

// A Wine the active Drinker has not tasted reads as not-yet-tasted.
func TestPreferences_WineVerdictUntastedIsNotTasted(t *testing.T) {
	wines := memory.NewWineRepo()
	tastings := memory.NewTastingRepo()
	varieties := memory.NewVarietyRepo()
	w, _ := domain.NewWine("Penfolds", "Bin 28 Shiraz", "Shiraz")
	wines.Save(w)
	sam, _ := domain.NewDrinker("Sam")

	h := app.NewPreferencesHandler(wines, varieties, tastings)
	got, err := h.WineVerdict(context.Background(), sam.ID, w.ID)
	if err != nil {
		t.Fatalf("WineVerdict: %v", err)
	}
	if got.Tasted {
		t.Errorf("verdict = %+v, want not-yet-tasted", got)
	}
}

// The Variety preference handler attributes enjoyment through the Compositions
// of the wines the active Drinker has drunk.
func TestPreferences_VarietyPreferenceThroughCompositions(t *testing.T) {
	wines := memory.NewWineRepo()
	tastings := memory.NewTastingRepo()
	varieties := memory.NewVarietyRepo()

	shiraz, _ := domain.NewVariety("Shiraz")
	varieties.Save(shiraz)
	w, _ := domain.NewWine("Penfolds", "Bin 28 Shiraz", "Shiraz")
	wines.Save(w)
	c, _ := domain.NewComposition([]domain.CompositionPart{{VarietyID: shiraz.ID, Proportion: 100}}, domain.ProvenanceConfirmed)
	_ = wines.SetComposition(context.Background(), w.ID, c)

	sam, _ := domain.NewDrinker("Sam")
	seedTasting(t, tastings, sam.ID, w.ID, 5)

	h := app.NewPreferencesHandler(wines, varieties, tastings)
	got, err := h.VarietyPreference(context.Background(), sam.ID, shiraz.ID)
	if err != nil {
		t.Fatalf("VarietyPreference: %v", err)
	}
	if !got.Tasted || got.Preference != 5 {
		t.Errorf("preference = %+v, want tasted, 5", got)
	}
}

// The Taste profile handler builds the enjoyed-grape set for the active
// Drinker, resolving Variety names for display, and keeps a multimodal palate
// as multiple grapes rather than averaging them away.
func TestPreferences_TasteProfileIsMultimodalSet(t *testing.T) {
	wines := memory.NewWineRepo()
	tastings := memory.NewTastingRepo()
	varieties := memory.NewVarietyRepo()

	shiraz, _ := domain.NewVariety("Shiraz")
	riesling, _ := domain.NewVariety("Riesling")
	varieties.Save(shiraz)
	varieties.Save(riesling)

	bigRed, _ := domain.NewWine("Penfolds", "Bin 28 Shiraz", "Shiraz")
	crispWhite, _ := domain.NewWine("Clare", "Riesling", "Riesling")
	wines.Save(bigRed)
	wines.Save(crispWhite)
	cR, _ := domain.NewComposition([]domain.CompositionPart{{VarietyID: shiraz.ID, Proportion: 100}}, domain.ProvenanceConfirmed)
	cW, _ := domain.NewComposition([]domain.CompositionPart{{VarietyID: riesling.ID, Proportion: 100}}, domain.ProvenanceConfirmed)
	_ = wines.SetComposition(context.Background(), bigRed.ID, cR)
	_ = wines.SetComposition(context.Background(), crispWhite.ID, cW)

	sam, _ := domain.NewDrinker("Sam")
	seedTasting(t, tastings, sam.ID, bigRed.ID, 5)
	seedTasting(t, tastings, sam.ID, crispWhite.ID, 5)

	h := app.NewPreferencesHandler(wines, varieties, tastings)
	profile, err := h.TasteProfile(context.Background(), sam.ID)
	if err != nil {
		t.Fatalf("TasteProfile: %v", err)
	}
	if len(profile.Enjoyed) != 2 {
		t.Fatalf("profile should keep both clusters; got %+v", profile)
	}
	names := map[string]bool{}
	for _, e := range profile.Enjoyed {
		names[e.Name] = true
	}
	if !names["Shiraz"] || !names["Riesling"] {
		t.Errorf("profile should name both enjoyed grapes; got %+v", profile.Enjoyed)
	}
}
