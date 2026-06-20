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

func TestDrinkersPage_ListsDrinkers(t *testing.T) {
	srv, _, drinkers := newDrinkerTestServer(t)
	partner, _ := domain.NewDrinker("Partner")
	if err := drinkers.Save(context.Background(), partner); err != nil {
		t.Fatalf("save drinker: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/drinkers", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Sam") || !strings.Contains(body, "Partner") {
		t.Errorf("page should list both Drinkers; got:\n%s", body)
	}
	// It is a full page (Layout shell), not a bare fragment.
	if !strings.Contains(body, "<html") {
		t.Errorf("GET /drinkers should render a full page; got:\n%s", body)
	}
	// The page hosts the add and rename forms (moved off the nav switcher).
	if !strings.Contains(body, `hx-post="/drinkers"`) {
		t.Errorf("page should host the add form; got:\n%s", body)
	}
	if !strings.Contains(body, `hx-put="/drinkers/`) {
		t.Errorf("page should host a rename form per Drinker; got:\n%s", body)
	}
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
	if loc := rec.Header().Get("Location"); loc != "/drinkers" {
		t.Errorf("Location = %q, want /drinkers", loc)
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
	body := rec.Body.String()
	if !strings.Contains(body, "name") {
		t.Errorf("failure should re-render the form with an error; got:\n%s", body)
	}
	// The 422 re-render targets the management region on the /drinkers page,
	// not the slimmed nav switcher.
	if !strings.Contains(body, `id="drinkers"`) {
		t.Errorf("failed add should re-render the #drinkers management region; got:\n%s", body)
	}
	if strings.Contains(body, `id="drinker-switcher"`) {
		t.Errorf("failed add must not re-render the nav switcher region; got:\n%s", body)
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
	if !strings.Contains(rec.Body.String(), `id="drinkers"`) {
		t.Errorf("failed rename should re-render the #drinkers management region; got:\n%s", rec.Body.String())
	}
	got, _ := drinkers.Get(context.Background(), d.ID)
	if got.Name != "Sam" {
		t.Errorf("a rejected rename must not persist: Name = %q, want Sam", got.Name)
	}
}
