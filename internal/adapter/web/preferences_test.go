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

// prefFixture wires a server over in-memory repos with one Drinker ("Sam"),
// returning the server and the repos so a test can seed grapes, wines and
// tastings. It is enough to exercise the do-I-like read models on the web layer.
type prefFixture struct {
	srv       *web.Server
	drinker   domain.Drinker
	wines     *memory.WineRepo
	varieties *memory.VarietyRepo
	tastings  *memory.TastingRepo
}

func newPrefFixture(t *testing.T) prefFixture {
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
	srv := web.NewServer(drinkers, wines, varieties, companions, logH, listH, listV, getV, editVC, listW, getW, editC, styleC, createD, renameD, prefs, app.NewDiscoveryHandler(wines, varieties, tastings))
	return prefFixture{srv: srv, drinker: d, wines: wines, varieties: varieties, tastings: tastings}
}

func (f prefFixture) logTasting(t *testing.T, wineID domain.ID, rating int) {
	t.Helper()
	r, _ := domain.NewRating(rating)
	ta, _ := domain.NewTasting(f.drinker.ID, wineID, nil, r, "", nil, time.Now())
	_ = f.tastings.Add(context.Background(), ta)
}

// The Wine detail page surfaces the active Drinker's verdict — "do I like this
// Wine?" — aggregated from their Tastings.
func TestWineDetail_ShowsVerdict(t *testing.T) {
	f := newPrefFixture(t)
	w, _ := domain.NewWine("Penfolds", "Bin 28 Shiraz", "Shiraz")
	f.wines.Save(w)
	f.logTasting(t, w.ID, 4)
	f.logTasting(t, w.ID, 5)

	req := httptest.NewRequest(http.MethodGet, "/wines/"+w.ID.String(), nil)
	rec := httptest.NewRecorder()
	f.srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body:\n%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	// The aggregate rating (mean 4.5 over 2 tastings) should be visible.
	if !strings.Contains(body, "4.5") {
		t.Errorf("wine detail should show the aggregate rating; got:\n%s", body)
	}
}

// An untasted Wine shows a clear not-yet-tasted message rather than a zero.
func TestWineDetail_UntastedShowsNotYetTasted(t *testing.T) {
	f := newPrefFixture(t)
	w, _ := domain.NewWine("Penfolds", "Bin 28 Shiraz", "Shiraz")
	f.wines.Save(w)

	req := httptest.NewRequest(http.MethodGet, "/wines/"+w.ID.String(), nil)
	rec := httptest.NewRecorder()
	f.srv.ServeHTTP(rec, req)

	body := strings.ToLower(rec.Body.String())
	if !strings.Contains(body, "not") || !strings.Contains(body, "tasted") {
		t.Errorf("untasted wine should show a not-yet-tasted message; got:\n%s", rec.Body.String())
	}
}

// The Variety detail page surfaces the active Drinker's derived preference,
// attributed through the Compositions of the wines drunk.
func TestVarietyDetail_ShowsDerivedPreference(t *testing.T) {
	f := newPrefFixture(t)
	shiraz, _ := domain.NewVariety("Shiraz")
	f.varieties.Save(shiraz)
	w, _ := domain.NewWine("Penfolds", "Bin 28 Shiraz", "Shiraz")
	f.wines.Save(w)
	c, _ := domain.NewComposition([]domain.CompositionPart{{VarietyID: shiraz.ID, Proportion: 100}}, domain.ProvenanceConfirmed)
	_ = f.wines.SetComposition(context.Background(), w.ID, c)
	f.logTasting(t, w.ID, 5)

	req := httptest.NewRequest(http.MethodGet, "/varieties/"+shiraz.ID.String(), nil)
	rec := httptest.NewRecorder()
	f.srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := strings.ToLower(rec.Body.String())
	// The derived preference should be surfaced as a "your preference" region.
	if !strings.Contains(body, "preference") {
		t.Errorf("variety detail should show the derived preference; got:\n%s", rec.Body.String())
	}
}

// The Varieties browse page surfaces the active Drinker's Taste profile — the
// SET of enjoyed grapes — and keeps a multimodal palate as multiple grapes.
func TestVarieties_ShowsTasteProfileAsSet(t *testing.T) {
	f := newPrefFixture(t)
	shiraz, _ := domain.NewVariety("Shiraz")
	riesling, _ := domain.NewVariety("Riesling")
	f.varieties.Save(shiraz)
	f.varieties.Save(riesling)

	bigRed, _ := domain.NewWine("Penfolds", "Bin 28 Shiraz", "Shiraz")
	crispWhite, _ := domain.NewWine("Clare", "Riesling", "Riesling")
	f.wines.Save(bigRed)
	f.wines.Save(crispWhite)
	cR, _ := domain.NewComposition([]domain.CompositionPart{{VarietyID: shiraz.ID, Proportion: 100}}, domain.ProvenanceConfirmed)
	cW, _ := domain.NewComposition([]domain.CompositionPart{{VarietyID: riesling.ID, Proportion: 100}}, domain.ProvenanceConfirmed)
	_ = f.wines.SetComposition(context.Background(), bigRed.ID, cR)
	_ = f.wines.SetComposition(context.Background(), crispWhite.ID, cW)
	f.logTasting(t, bigRed.ID, 5)
	f.logTasting(t, crispWhite.ID, 5)

	req := httptest.NewRequest(http.MethodGet, "/varieties", nil)
	rec := httptest.NewRecorder()
	f.srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(strings.ToLower(body), "taste profile") {
		t.Errorf("varieties page should show the Taste profile; got:\n%s", body)
	}
	// Both clusters — the bold red and the crisp white — survive as a set.
	if !strings.Contains(body, "Shiraz") || !strings.Contains(body, "Riesling") {
		t.Errorf("multimodal Taste profile should list both enjoyed grapes; got:\n%s", body)
	}
}
