package app_test

import (
	"context"
	"testing"
	"time"

	"go-wine/internal/adapter/memory"
	"go-wine/internal/app"
	"go-wine/internal/domain"
)

// TestLogTasting_AttachesCompanions logs a Tasting with Companions and confirms
// the read-side view surfaces their names. This is the behaviour the feature
// exists for, exercised end-to-end over the repository ports.
func TestLogTasting_AttachesCompanions(t *testing.T) {
	drinkers, wines, tastings, d, w := fixture(t)
	companions := memory.NewCompanionRepo()

	alex, _ := domain.NewCompanion(d.ID, "Alex")
	jo, _ := domain.NewCompanion(d.ID, "Jo")
	for _, c := range []domain.Companion{alex, jo} {
		if err := companions.Add(context.Background(), c); err != nil {
			t.Fatalf("add companion: %v", err)
		}
	}

	log := app.NewLogTastingHandler(drinkers, wines, tastings)
	if _, err := log.Handle(context.Background(), app.LogTastingCommand{
		DrinkerID:  d.ID,
		WineID:     w.ID,
		Rating:     4,
		Companions: []domain.ID{alex.ID, jo.ID},
		DrunkOn:    time.Now(),
	}); err != nil {
		t.Fatalf("log: %v", err)
	}

	q := app.NewListTastingsHandler(wines, tastings, companions)
	views, err := q.Handle(context.Background(), d.ID)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if len(views) != 1 {
		t.Fatalf("expected 1 view, got %d", len(views))
	}
	got := views[0].Companions
	if len(got) != 2 {
		t.Fatalf("view carries %d companions, want 2: %v", len(got), got)
	}
	names := map[string]bool{got[0]: true, got[1]: true}
	if !names["Alex"] || !names["Jo"] {
		t.Errorf("companions = %v, want Alex and Jo", got)
	}
}
