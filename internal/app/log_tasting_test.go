package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"go-wine/internal/adapter/memory"
	"go-wine/internal/app"
	"go-wine/internal/domain"
)

func fixture(t *testing.T) (*memory.DrinkerRepo, *memory.WineRepo, *memory.TastingRepo, domain.Drinker, domain.Wine) {
	t.Helper()
	drinkers := memory.NewDrinkerRepo()
	wines := memory.NewWineRepo()
	tastings := memory.NewTastingRepo()

	d, err := domain.NewDrinker("Sam")
	if err != nil {
		t.Fatalf("new drinker: %v", err)
	}
	drinkers.Save(d)

	w, err := domain.NewWine("Penfolds", "Bin 28 Shiraz", "Shiraz")
	if err != nil {
		t.Fatalf("new wine: %v", err)
	}
	wines.Save(w)

	return drinkers, wines, tastings, d, w
}

func TestLogTasting_Succeeds(t *testing.T) {
	drinkers, wines, tastings, d, w := fixture(t)
	h := app.NewLogTastingHandler(drinkers, wines, tastings)

	id, err := h.Handle(context.Background(), app.LogTastingCommand{
		DrinkerID: d.ID,
		WineID:    w.ID,
		Rating:    4,
		Note:      "cold rainy night with the lamb stew",
		DrunkOn:   time.Date(2026, 6, 15, 20, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id == "" {
		t.Fatal("expected a non-empty tasting id")
	}

	got, err := tastings.ListByDrinker(context.Background(), d.ID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 tasting, got %d", len(got))
	}
	if got[0].Rating.Int() != 4 {
		t.Errorf("rating = %d, want 4", got[0].Rating.Int())
	}
}

func TestLogTasting_RejectsOutOfRangeRating(t *testing.T) {
	drinkers, wines, tastings, d, w := fixture(t)
	h := app.NewLogTastingHandler(drinkers, wines, tastings)

	_, err := h.Handle(context.Background(), app.LogTastingCommand{
		DrinkerID: d.ID,
		WineID:    w.ID,
		Rating:    6,
		DrunkOn:   time.Now(),
	})
	if !errors.Is(err, domain.ErrInvalidRating) {
		t.Fatalf("err = %v, want ErrInvalidRating", err)
	}
}

func TestLogTasting_RejectsUnknownWine(t *testing.T) {
	drinkers, wines, tastings, d, _ := fixture(t)
	h := app.NewLogTastingHandler(drinkers, wines, tastings)

	_, err := h.Handle(context.Background(), app.LogTastingCommand{
		DrinkerID: d.ID,
		WineID:    domain.NewID(),
		Rating:    3,
		DrunkOn:   time.Now(),
	})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestLogTasting_RejectsUnknownDrinker(t *testing.T) {
	drinkers, wines, tastings, _, w := fixture(t)
	h := app.NewLogTastingHandler(drinkers, wines, tastings)

	_, err := h.Handle(context.Background(), app.LogTastingCommand{
		DrinkerID: domain.NewID(),
		WineID:    w.ID,
		Rating:    3,
		DrunkOn:   time.Now(),
	})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestListTastings_MostRecentFirst(t *testing.T) {
	drinkers, wines, tastings, d, w := fixture(t)
	log := app.NewLogTastingHandler(drinkers, wines, tastings)

	older := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	newer := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	for _, when := range []time.Time{older, newer} {
		if _, err := log.Handle(context.Background(), app.LogTastingCommand{
			DrinkerID: d.ID, WineID: w.ID, Rating: 3, DrunkOn: when,
		}); err != nil {
			t.Fatalf("log: %v", err)
		}
	}

	q := app.NewListTastingsHandler(wines, tastings)
	views, err := q.Handle(context.Background(), d.ID)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if len(views) != 2 {
		t.Fatalf("expected 2 views, got %d", len(views))
	}
	if !views[0].DrunkOn.Equal(newer) {
		t.Errorf("first view = %v, want newest %v", views[0].DrunkOn, newer)
	}
	if views[0].WineLabel != "Penfolds — Bin 28 Shiraz" {
		t.Errorf("label = %q", views[0].WineLabel)
	}
}
