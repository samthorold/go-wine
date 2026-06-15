package web_test

import (
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

// newTestServer wires the web adapter against in-memory repositories with a
// single Drinker and a single Wine, returning the server, that Drinker, and
// that Wine.
func newTestServer(t *testing.T) (*web.Server, domain.Drinker, domain.Wine) {
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

	logH := app.NewLogTastingHandler(drinkers, wines, tastings)
	listH := app.NewListTastingsHandler(wines, tastings)
	listV := app.NewListVarietiesHandler(memory.NewVarietyRepo())
	return web.NewServer(drinkers, wines, logH, listH, listV), d, w
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
