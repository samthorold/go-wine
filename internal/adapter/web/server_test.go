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
	"go-wine/internal/seed"
)

// newTestServer wires the web adapter against in-memory repositories with a
// single Drinker and a single Wine, returning the server, that Drinker, and
// that Wine.
func newTestServer(t *testing.T) (*web.Server, domain.Drinker, domain.Wine) {
	srv, d, w, _ := newTestServerWithCompanions(t)
	return srv, d, w
}

// newTestServerWithCompanions also exposes the in-memory CompanionRepo so tests
// can seed existing Companions and assert that new ones get persisted.
func newTestServerWithCompanions(t *testing.T) (*web.Server, domain.Drinker, domain.Wine, *memory.CompanionRepo) {
	t.Helper()
	drinkers := memory.NewDrinkerRepo()
	wines := memory.NewWineRepo()
	tastings := memory.NewTastingRepo()
	varieties := memory.NewVarietyRepo()
	companions := memory.NewCompanionRepo()

	d, err := domain.NewDrinker("Sam")
	if err != nil {
		t.Fatalf("new drinker: %v", err)
	}
	if err := drinkers.Save(context.Background(), d); err != nil {
		t.Fatalf("save drinker: %v", err)
	}

	w, err := domain.NewWine("Penfolds", "Bin 28 Shiraz", "Shiraz")
	if err != nil {
		t.Fatalf("new wine: %v", err)
	}
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
	return web.NewServer(drinkers, wines, varieties, companions, logH, listH, listV, getV, editVC, listW, getW, editC, styleC, createD, renameD, prefs, app.NewDiscoveryHandler(wines, varieties, tastings)), d, w, companions
}

func TestSwitch_PostSetsCookieAndRedirectsToTastings(t *testing.T) {
	srv, d, _ := newTestServer(t)

	form := url.Values{"drinker": {d.ID.String()}}
	req := httptest.NewRequest(http.MethodPost, "/switch", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusSeeOther)
	}
	if loc := rec.Header().Get("Location"); loc != "/tastings" {
		t.Errorf("Location = %q, want /tastings", loc)
	}

	var got *http.Cookie
	for _, c := range rec.Result().Cookies() {
		if c.Name == "drinker" {
			got = c
		}
	}
	if got == nil {
		t.Fatalf("expected a drinker cookie to be set")
	}
	if got.Value != d.ID.String() {
		t.Errorf("cookie value = %q, want %q", got.Value, d.ID.String())
	}
}

func TestSwitch_GetIsNotRouted(t *testing.T) {
	srv, d, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/switch?drinker="+d.ID.String(), nil)
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if rec.Code == http.StatusSeeOther || rec.Code == http.StatusOK {
		t.Fatalf("GET /switch should not be routed (no safe-method mutation); got status %d", rec.Code)
	}
	for _, c := range rec.Result().Cookies() {
		if c.Name == "drinker" {
			t.Errorf("GET /switch must not set a drinker cookie")
		}
	}
}

func postTasting(t *testing.T, srv *web.Server, form url.Values) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/tastings", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	return rec
}

func TestLogTasting_SuccessSwapsFormAndOOBList(t *testing.T) {
	srv, _, w := newTestServer(t)

	rec := postTasting(t, srv, url.Values{
		"wine_id": {w.ID.String()},
		"rating":  {"4"},
		"note":    {"lamb stew"},
	})

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `id="log-form"`) {
		t.Errorf("success should return a fresh log form; got:\n%s", body)
	}
	if !strings.Contains(body, `hx-swap-oob="true"`) {
		t.Errorf("success should include the OOB tastings list; got:\n%s", body)
	}
	if !strings.Contains(body, "Bin 28 Shiraz") {
		t.Errorf("OOB list should contain the newly logged tasting; got:\n%s", body)
	}
	// The fresh form must be reset: the entered note appears only in the OOB
	// list row, never re-populated into the textarea.
	if strings.Contains(body, `cosy...">lamb stew`) {
		t.Errorf("fresh form should be empty, note must not be preserved; got:\n%s", body)
	}
}

func TestLogTasting_PicksExistingAndAddsNewCompanions(t *testing.T) {
	srv, d, w, companions := newTestServerWithCompanions(t)

	alex, _ := domain.NewCompanion(d.ID, "Alex")
	if err := companions.Add(context.Background(), alex); err != nil {
		t.Fatalf("seed companion: %v", err)
	}

	rec := postTasting(t, srv, url.Values{
		"wine_id":        {w.ID.String()},
		"rating":         {"4"},
		"companion_id":   {alex.ID.String()},
		"new_companions": {"Jo"},
	})

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Alex") || !strings.Contains(body, "Jo") {
		t.Errorf("OOB list should show both the picked and the new companion; got:\n%s", body)
	}

	// The new Companion is persisted in the Drinker's personal zone.
	cs, err := companions.ListByDrinker(context.Background(), d.ID)
	if err != nil {
		t.Fatalf("list companions: %v", err)
	}
	var names []string
	for _, c := range cs {
		names = append(names, c.Name)
	}
	if len(cs) != 2 {
		t.Fatalf("expected Alex + Jo persisted, got %v", names)
	}
}

func TestLogTasting_BadRatingReRendersFormWith422(t *testing.T) {
	srv, _, w := newTestServer(t)

	rec := postTasting(t, srv, url.Values{
		"wine_id": {w.ID.String()},
		"rating":  {"6"},
		"note":    {"too good to rate"},
	})

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnprocessableEntity)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `id="log-form"`) {
		t.Errorf("failure should re-render the form; got:\n%s", body)
	}
	if !strings.Contains(body, "rating must be between 1 and 5") {
		t.Errorf("failure should show the inline rating error; got:\n%s", body)
	}
	if !strings.Contains(body, "too good to rate") {
		t.Errorf("failure should preserve the entered note; got:\n%s", body)
	}
	if strings.Contains(body, `hx-swap-oob="true"`) {
		t.Errorf("failure must not touch the tastings list (no OOB); got:\n%s", body)
	}
}

func TestLogTasting_NoRatingReRendersFormWith422(t *testing.T) {
	// A fresh tasting starts unrated; submitting it with the no-rating option
	// (an empty "rating") is rejected by the domain (the authority), re-rendering
	// the form with an inline error and the entered values preserved — not saved
	// as a phantom rating. See issue #39.
	srv, _, w := newTestServer(t)

	rec := postTasting(t, srv, url.Values{
		"wine_id": {w.ID.String()},
		"rating":  {""},
		"note":    {"forgot to rate"},
	})

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnprocessableEntity)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `id="log-form"`) {
		t.Errorf("no-rating submit should re-render the form; got:\n%s", body)
	}
	if !strings.Contains(body, "rating must be between 1 and 5") {
		t.Errorf("no-rating submit should show the inline rating error; got:\n%s", body)
	}
	if !strings.Contains(body, "forgot to rate") {
		t.Errorf("no-rating submit should preserve the entered note; got:\n%s", body)
	}
	if strings.Contains(body, `hx-swap-oob="true"`) {
		t.Errorf("no-rating failure must not touch the tastings list (no OOB); got:\n%s", body)
	}
}

func TestLogTasting_UnknownWineReRendersFormWith422(t *testing.T) {
	srv, _, _ := newTestServer(t)

	rec := postTasting(t, srv, url.Values{
		"wine_id": {domain.NewID().String()},
		"rating":  {"3"},
	})

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnprocessableEntity)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `id="log-form"`) {
		t.Errorf("failure should re-render the form; got:\n%s", body)
	}
}
