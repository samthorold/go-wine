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
// single Drinker, returning the server and that Drinker.
func newTestServer(t *testing.T) (*web.Server, domain.Drinker) {
	t.Helper()
	drinkers := memory.NewDrinkerRepo()
	wines := memory.NewWineRepo()
	tastings := memory.NewTastingRepo()

	d, err := domain.NewDrinker("Sam")
	if err != nil {
		t.Fatalf("new drinker: %v", err)
	}
	drinkers.Save(d)

	logH := app.NewLogTastingHandler(drinkers, wines, tastings)
	listH := app.NewListTastingsHandler(wines, tastings)
	return web.NewServer(drinkers, wines, logH, listH), d
}

func TestSwitch_PostSetsCookieAndRedirectsToTastings(t *testing.T) {
	srv, d := newTestServer(t)

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
	srv, d := newTestServer(t)

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
