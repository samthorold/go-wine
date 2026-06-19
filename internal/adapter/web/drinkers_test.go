package web_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"go-wine/internal/adapter/memory"
	"go-wine/internal/adapter/web"
	"go-wine/internal/app"
	"go-wine/internal/domain"
)

// newDrinkerTestServer wires the web adapter against in-memory repositories with
// a single seeded Drinker, returning the server, that Drinker, and the Drinker
// repo so tests can assert persistence.
func newDrinkerTestServer(t *testing.T) (*web.Server, domain.Drinker, *memory.DrinkerRepo) {
	t.Helper()
	drinkers := memory.NewDrinkerRepo()
	wines := memory.NewWineRepo()
	tastings := memory.NewTastingRepo()
	companions := memory.NewCompanionRepo()

	d, _ := domain.NewDrinker("Sam")
	if err := drinkers.Save(context.Background(), d); err != nil {
		t.Fatalf("save drinker: %v", err)
	}

	varieties := memory.NewVarietyRepo()
	logH := app.NewLogTastingHandler(drinkers, wines, tastings)
	listH := app.NewListTastingsHandler(wines, tastings, companions)
	listV := app.NewListVarietiesHandler(varieties)
	getV := app.NewGetVarietyHandler(varieties)
	editVC := app.NewEditCharacteristicsHandler(varieties)
	listW := app.NewListWinesHandler(wines)
	getW := app.NewGetWineHandler(wines, varieties)
	editC := app.NewEditCompositionHandler(wines, varieties)
	styleC := app.NewResolveStyleCompositionHandler(varieties, nil)
	createH := app.NewCreateDrinkerHandler(drinkers)
	renameH := app.NewRenameDrinkerHandler(drinkers)
	prefs := app.NewPreferencesHandler(wines, varieties, tastings)
	srv := web.NewServer(drinkers, wines, varieties, companions, logH, listH, listV, getV, editVC, listW, getW, editC, styleC, createH, renameH, prefs, app.NewDiscoveryHandler(wines, varieties, tastings))
	return srv, d, drinkers
}

func TestCreateDrinker_PostCreatesAndSwitchesTo(t *testing.T) {
	srv, _, drinkers := newDrinkerTestServer(t)

	form := url.Values{"name": {"Partner"}}
	req := httptest.NewRequest(http.MethodPost, "/drinkers", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusSeeOther)
	}
	if loc := rec.Header().Get("Location"); loc != "/tastings" {
		t.Errorf("Location = %q, want /tastings", loc)
	}

	// The new Drinker is persisted and selectable.
	all, _ := drinkers.List(context.Background())
	var partner domain.Drinker
	for _, d := range all {
		if d.Name == "Partner" {
			partner = d
		}
	}
	if partner.ID == "" {
		t.Fatalf("Partner was not persisted; got %v", all)
	}

	// And becomes the active Drinker via the switcher cookie.
	var cookie *http.Cookie
	for _, c := range rec.Result().Cookies() {
		if c.Name == "drinker" {
			cookie = c
		}
	}
	if cookie == nil || cookie.Value != partner.ID.String() {
		t.Errorf("new Drinker should become active: cookie = %v, want %q", cookie, partner.ID)
	}
}

func TestCreateDrinker_EmptyNameReRendersWith422(t *testing.T) {
	srv, _, _ := newDrinkerTestServer(t)

	form := url.Values{"name": {""}}
	req := httptest.NewRequest(http.MethodPost, "/drinkers", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnprocessableEntity)
	}
	if !strings.Contains(rec.Body.String(), "name") {
		t.Errorf("failure should re-render the form with an error; got:\n%s", rec.Body.String())
	}
}

func TestRenameDrinker_PutRenamesAndRedirects(t *testing.T) {
	srv, d, drinkers := newDrinkerTestServer(t)

	form := url.Values{"name": {"Samuel"}}
	req := httptest.NewRequest(http.MethodPut, "/drinkers/"+d.ID.String(), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusSeeOther)
	}
	got, _ := drinkers.Get(context.Background(), d.ID)
	if got.Name != "Samuel" {
		t.Errorf("rename did not persist: Name = %q, want Samuel", got.Name)
	}
}

func TestRenameDrinker_EmptyNameReRendersWith422(t *testing.T) {
	srv, d, drinkers := newDrinkerTestServer(t)

	form := url.Values{"name": {""}}
	req := httptest.NewRequest(http.MethodPut, "/drinkers/"+d.ID.String(), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnprocessableEntity)
	}
	got, _ := drinkers.Get(context.Background(), d.ID)
	if got.Name != "Sam" {
		t.Errorf("a rejected rename must not persist: Name = %q, want Sam", got.Name)
	}
}
