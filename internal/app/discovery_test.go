package app_test

import (
	"context"
	"testing"
	"time"

	"go-wine/internal/adapter/memory"
	"go-wine/internal/app"
	"go-wine/internal/domain"
)

// discoveryFixture wires the in-memory repos and seeds a Variety with
// characteristics, returning its ID. Wines/Tastings are added per-test.
type discoveryFixture struct {
	wines     *memory.WineRepo
	varieties *memory.VarietyRepo
	tastings  *memory.TastingRepo
}

func newDiscoveryFixture() discoveryFixture {
	return discoveryFixture{
		wines:     memory.NewWineRepo(),
		varieties: memory.NewVarietyRepo(),
		tastings:  memory.NewTastingRepo(),
	}
}

func (f discoveryFixture) variety(name string, body, tannin, acidity, sweetness, alcohol int, notes ...string) domain.Variety {
	v, _ := domain.NewVariety(name)
	f.varieties.Save(v)
	c, _ := domain.NewCharacteristics(body, tannin, acidity, sweetness, alcohol, notes, domain.ProvenanceDefault)
	_ = f.varieties.SetCharacteristics(context.Background(), v.ID, c)
	return v
}

// singleGrapeWine creates a Wine composed 100% of the given Variety.
func (f discoveryFixture) singleGrapeWine(label string, v domain.Variety) domain.Wine {
	w, _ := domain.NewWine("Producer", label, "")
	f.wines.Save(w)
	c, _ := domain.NewComposition([]domain.CompositionPart{{VarietyID: v.ID, Proportion: 100}}, domain.ProvenanceConfirmed)
	_ = f.wines.SetComposition(context.Background(), w.ID, c)
	return w
}

func (f discoveryFixture) logTasting(drinkerID, wineID domain.ID, rating int) {
	r, _ := domain.NewRating(rating)
	ta, _ := domain.NewTasting(drinkerID, wineID, nil, r, "", nil, time.Now())
	_ = f.tastings.Add(context.Background(), ta)
}

// The Discovery handler recommends untried Varieties near the active Drinker's
// enjoyed-grape set, naming the enjoyed grape(s) that justify each one, and
// never recommends a grape the Drinker has already tried.
func TestDiscovery_RecommendsUntriedNearProfileWithExplanation(t *testing.T) {
	f := newDiscoveryFixture()
	sam, _ := domain.NewDrinker("Sam")

	nebbiolo := f.variety("Nebbiolo", 5, 5, 4, 1, 4, "cherry", "tar")
	aglianico := f.variety("Aglianico", 5, 5, 4, 1, 4, "cherry", "leather") // near Nebbiolo, untried
	pinotGrigio := f.variety("Pinot Grigio", 1, 1, 4, 1, 1)                 // far, untried

	// Sam loved a Nebbiolo wine — Nebbiolo enters the profile and is "tried".
	w := f.singleGrapeWine("Barolo", nebbiolo)
	f.logTasting(sam.ID, w.ID, 5)

	h := app.NewDiscoveryHandler(f.wines, f.varieties, f.tastings)
	recs, err := h.Handle(context.Background(), sam.ID)
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}

	if len(recs) == 0 {
		t.Fatalf("expected recommendations")
	}
	// Nebbiolo is tried — it must never be recommended.
	for _, r := range recs {
		if r.VarietyID == nebbiolo.ID {
			t.Fatalf("tried grape Nebbiolo must not be recommended; got %+v", recs)
		}
	}
	// The nearest untried grape ranks first and is explained by Nebbiolo (by name).
	if recs[0].VarietyID != aglianico.ID {
		t.Errorf("nearest untried grape should rank first; got %q", recs[0].Name)
	}
	if recs[0].Name != "Aglianico" {
		t.Errorf("recommendation should carry the resolved name; got %q", recs[0].Name)
	}
	if len(recs[0].Because) == 0 || recs[0].Because[0] != "Nebbiolo" {
		t.Errorf("recommendation should name the justifying enjoyed grape; got %+v", recs[0].Because)
	}
	// The far white still appears (all ranked) but below Aglianico.
	if recs[len(recs)-1].VarietyID != pinotGrigio.ID {
		t.Errorf("far grape should rank last; got %q", recs[len(recs)-1].Name)
	}
}

// A Drinker with no enjoyed grapes gets an empty recommendation set — the page
// renders an explanatory empty state rather than an arbitrary list.
func TestDiscovery_EmptyProfileYieldsNoRecommendations(t *testing.T) {
	f := newDiscoveryFixture()
	sam, _ := domain.NewDrinker("Sam")
	f.variety("Aglianico", 5, 5, 4, 1, 4)

	h := app.NewDiscoveryHandler(f.wines, f.varieties, f.tastings)
	recs, err := h.Handle(context.Background(), sam.ID)
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if len(recs) != 0 {
		t.Errorf("empty profile should yield no recommendations; got %+v", recs)
	}
}
