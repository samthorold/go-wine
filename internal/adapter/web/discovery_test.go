package web_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go-wine/internal/adapter/memory"
	"go-wine/internal/adapter/web"
	"go-wine/internal/app"
	"go-wine/internal/domain"
)

// discoFixture wires a server over in-memory repos with one Drinker ("Sam"),
// returning the server and repos so a test can seed grapes, characteristics,
// wines and tastings to exercise the Discovery page end to end.
type discoFixture struct {
	srv       *web.Server
	drinker   domain.Drinker
	wines     *memory.WineRepo
	varieties *memory.VarietyRepo
	tastings  *memory.TastingRepo
}

func newDiscoFixture(t *testing.T) discoFixture {
	t.Helper()
	drinkers := memory.NewDrinkerRepo()
	wines := memory.NewWineRepo()
	tastings := memory.NewTastingRepo()
	varieties := memory.NewVarietyRepo()
	companions := memory.NewCompanionRepo()

	d, _ := domain.NewDrinker("Sam")
	_ = drinkers.Save(context.Background(), d)

	logH := app.NewLogTastingHandler(drinkers, wines, tastings)
	listH := app.NewListTastingsHandler(wines, tastings, companions)
	listV := app.NewListVarietiesHandler(varieties)
	getV := app.NewGetVarietyHandler(varieties)
	editVC := app.NewEditCharacteristicsHandler(varieties)
	listW := app.NewListWinesHandler(wines)
	getW := app.NewGetWineHandler(wines, varieties)
	editC := app.NewEditCompositionHandler(wines, varieties)
	styleC := app.NewResolveStyleCompositionHandler(varieties, nil)
	createD := app.NewCreateDrinkerHandler(drinkers)
	renameD := app.NewRenameDrinkerHandler(drinkers)
	prefs := app.NewPreferencesHandler(wines, varieties, tastings)
	disco := app.NewDiscoveryHandler(wines, varieties, tastings)
	srv := web.NewServer(drinkers, wines, varieties, companions, logH, listH, listV, getV, editVC, listW, getW, editC, styleC, createD, renameD, prefs, disco)
	return discoFixture{srv: srv, drinker: d, wines: wines, varieties: varieties, tastings: tastings}
}

func (f discoFixture) variety(name string, body, tannin, acidity, sweetness, alcohol int, notes ...string) domain.Variety {
	v, _ := domain.NewVariety(name)
	f.varieties.Save(v)
	c, _ := domain.NewCharacteristics(body, tannin, acidity, sweetness, alcohol, notes, domain.ProvenanceDefault)
	_ = f.varieties.SetCharacteristics(context.Background(), v.ID, c)
	return v
}

func (f discoFixture) wineOf(label string, v domain.Variety) domain.Wine {
	w, _ := domain.NewWine("Producer", label, "")
	f.wines.Save(w)
	c, _ := domain.NewComposition([]domain.CompositionPart{{VarietyID: v.ID, Proportion: 100}}, domain.ProvenanceConfirmed)
	_ = f.wines.SetComposition(context.Background(), w.ID, c)
	return w
}

func (f discoFixture) logTasting(wineID domain.ID, rating int) {
	r, _ := domain.NewRating(rating)
	ta, _ := domain.NewTasting(f.drinker.ID, wineID, nil, r, "", nil, time.Now())
	_ = f.tastings.Add(context.Background(), ta)
}

// The Discovery page shows recommendations for the active Drinker, naming the
// enjoyed grape that justifies each, and never recommends a tried grape.
func TestDiscoveryPage_ShowsRecommendationsForActiveDrinker(t *testing.T) {
	f := newDiscoFixture(t)
	nebbiolo := f.variety("Nebbiolo", 5, 5, 4, 1, 4, "cherry", "tar")
	f.variety("Aglianico", 5, 5, 4, 1, 4, "cherry", "leather") // near, untried
	w := f.wineOf("Barolo", nebbiolo)
	f.logTasting(w.ID, 5)

	req := httptest.NewRequest(http.MethodGet, "/discovery", nil)
	rec := httptest.NewRecorder()
	f.srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body:\n%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Aglianico") {
		t.Errorf("discovery should recommend the near untried grape; got:\n%s", body)
	}
	if !strings.Contains(body, "Nebbiolo") {
		t.Errorf("recommendation should name the justifying enjoyed grape; got:\n%s", body)
	}
}

// With no enjoyed grapes, the Discovery page renders an explanatory empty state.
func TestDiscoveryPage_EmptyProfile(t *testing.T) {
	f := newDiscoFixture(t)
	f.variety("Aglianico", 5, 5, 4, 1, 4)

	req := httptest.NewRequest(http.MethodGet, "/discovery", nil)
	rec := httptest.NewRecorder()
	f.srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(strings.ToLower(rec.Body.String()), "no recommendations") {
		t.Errorf("empty profile should show an explanatory empty state; got:\n%s", rec.Body.String())
	}
}
