package web_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go-wine/internal/adapter/memory"
	"go-wine/internal/adapter/web"
	"go-wine/internal/app"
	"go-wine/internal/domain"
	"go-wine/internal/seed"
)

// newWineServer wires the web adapter against in-memory stores seeded with the
// given grapes and a single Wine of the given Style, returning the server and the
// Wine. It is enough to exercise the Style → default Composition fill flow.
func newWineServer(t *testing.T, style string, grapes ...string) (*web.Server, domain.Wine, *memory.WineRepo, map[string]domain.ID) {
	t.Helper()
	drinkers := memory.NewDrinkerRepo()
	wines := memory.NewWineRepo()
	tastings := memory.NewTastingRepo()
	varieties := memory.NewVarietyRepo()
	companions := memory.NewCompanionRepo()

	d, _ := domain.NewDrinker("Sam")
	_ = drinkers.Save(context.Background(), d)

	byName := make(map[string]domain.ID, len(grapes))
	for _, g := range grapes {
		v, _ := domain.NewVariety(g)
		varieties.Save(v)
		byName[g] = v.ID
	}

	w, _ := domain.NewWine("Guigal", "Côtes du Rhône Rouge", style)
	wines.Save(w)

	logH := app.NewLogTastingHandler(drinkers, wines, tastings)
	listH := app.NewListTastingsHandler(wines, tastings, companions)
	listV := app.NewListVarietiesHandler(varieties)
	getV := app.NewGetVarietyHandler(varieties)
	editVC := app.NewEditCharacteristicsHandler(varieties)
	listW := app.NewListWinesHandler(wines)
	getW := app.NewGetWineHandler(wines, varieties)
	editC := app.NewEditCompositionHandler(wines, varieties)
	styleC := app.NewResolveStyleCompositionHandler(varieties, seed.StyleCompositions())
	createD := app.NewCreateDrinkerHandler(drinkers)
	renameD := app.NewRenameDrinkerHandler(drinkers)
	prefs := app.NewPreferencesHandler(wines, varieties, tastings)
	srv := web.NewServer(drinkers, wines, varieties, companions, logH, listH, listV, getV, editVC, listW, getW, editC, styleC, createD, renameD, prefs, app.NewDiscoveryHandler(wines, varieties, tastings))
	return srv, w, wines, byName
}

// Asking for a Wine's Style default returns the Composition form prefilled with
// the conventional grapes for that Style, ready for the Drinker to override.
func TestWineStyleDefault_PrefillsFormFromStyle(t *testing.T) {
	srv, w, _, byName := newWineServer(t, "GSM", "Grenache", "Syrah", "Mourvèdre")

	req := httptest.NewRequest(http.MethodGet, "/wines/"+w.ID.String()+"/style-default", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `id="composition-form"`) {
		t.Errorf("should return the composition form fragment; got:\n%s", body)
	}
	// Each GSM grape is pre-selected in a row.
	for _, g := range []string{"Grenache", "Syrah", "Mourvèdre"} {
		if !strings.Contains(body, `value="`+byName[g].String()+`" selected`) {
			t.Errorf("form should pre-select %s from the GSM default; got:\n%s", g, body)
		}
	}
	// Grenache leads at 50%.
	if !strings.Contains(body, `value="50"`) {
		t.Errorf("form should carry the GSM default proportions; got:\n%s", body)
	}
}

// Saving the prefilled (or overridden) grapes confirms the Composition, so a
// later Style re-seed cannot clobber it.
func TestWineStyleDefault_SaveConfirmsComposition(t *testing.T) {
	srv, w, wines, byName := newWineServer(t, "GSM", "Grenache", "Syrah", "Mourvèdre")

	form := strings.NewReader(
		"variety_id=" + byName["Grenache"].String() + "&proportion=100")
	req := httptest.NewRequest(http.MethodPut, "/wines/"+w.ID.String(), form)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body:\n%s", rec.Code, rec.Body.String())
	}
	got, _ := wines.Get(context.Background(), w.ID)
	if !got.Composition.IsConfirmed() {
		t.Errorf("a hand-saved Composition should be confirmed; got %+v", got.Composition)
	}
}

// A Style with no conventional default returns the form with a gentle banner
// rather than an error.
func TestWineStyleDefault_UnknownStyleShowsBanner(t *testing.T) {
	srv, w, _, _ := newWineServer(t, "Mystery", "Grenache")

	req := httptest.NewRequest(http.MethodGet, "/wines/"+w.ID.String()+"/style-default", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(strings.ToLower(rec.Body.String()), "no conventional grapes") {
		t.Errorf("unknown style should show a gentle banner; got:\n%s", rec.Body.String())
	}
}
