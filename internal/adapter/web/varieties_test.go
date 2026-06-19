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

// newVarietyTestServer wires a server over in-memory repos with one Drinker and
// returns it plus the variety repo so tests can seed grapes.
func newVarietyTestServer(t *testing.T) (*web.Server, *memory.VarietyRepo) {
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
	srv := web.NewServer(drinkers, wines, varieties, companions, logH, listH, listV, getV, editVC, listW, getW, editC, styleC, createD, renameD, prefs)
	return srv, varieties
}

func TestVarieties_GetListsSeededVarieties(t *testing.T) {
	srv, varieties := newVarietyTestServer(t)
	v, _ := domain.NewVariety("Shiraz")
	varieties.Save(v)

	req := httptest.NewRequest(http.MethodGet, "/varieties", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "Shiraz") {
		t.Errorf("GET /varieties should list the seeded Variety; got:\n%s", rec.Body.String())
	}
}

func TestVariety_GetDetailShowsCharacteristics(t *testing.T) {
	srv, varieties := newVarietyTestServer(t)
	v, _ := domain.NewVariety("Shiraz")
	varieties.Save(v)
	c, _ := domain.NewCharacteristics(5, 4, 2, 1, 5, []string{"blackberry"}, domain.ProvenanceDefault)
	_ = varieties.SetCharacteristics(context.Background(), v.ID, c)

	req := httptest.NewRequest(http.MethodGet, "/varieties/"+v.ID.String(), nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Shiraz") || !strings.Contains(body, "blackberry") {
		t.Errorf("detail should show name and notes; got:\n%s", body)
	}
}

func TestVariety_GetUnknownIs404(t *testing.T) {
	srv, _ := newVarietyTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/varieties/nope", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestVariety_PutValidEditConfirmsAndUpdatesView(t *testing.T) {
	srv, varieties := newVarietyTestServer(t)
	v, _ := domain.NewVariety("Shiraz")
	varieties.Save(v)

	form := url.Values{
		"body": {"5"}, "tannin": {"4"}, "acidity": {"2"}, "sweetness": {"1"}, "alcohol": {"5"},
		"notes": {"blackberry, pepper"},
	}
	req := httptest.NewRequest(http.MethodPut, "/varieties/"+v.ID.String(), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body:\n%s", rec.Code, rec.Body.String())
	}
	// The edit confirmed the bundle.
	got, _ := varieties.GetCharacteristics(context.Background(), v.ID)
	if !got.IsConfirmed() || got.Body.Int() != 5 {
		t.Errorf("PUT should persist confirmed characteristics; got %+v", got)
	}
	// The response carries the updated OOB characteristics view.
	if !strings.Contains(rec.Body.String(), "confirmed") {
		t.Errorf("response should reflect the now-confirmed provenance; got:\n%s", rec.Body.String())
	}
}

func TestVariety_PutOutOfScaleIs422AndPreservesValues(t *testing.T) {
	srv, varieties := newVarietyTestServer(t)
	v, _ := domain.NewVariety("Shiraz")
	varieties.Save(v)

	form := url.Values{
		"body": {"9"}, "tannin": {"4"}, "acidity": {"2"}, "sweetness": {"1"}, "alcohol": {"5"},
	}
	req := httptest.NewRequest(http.MethodPut, "/varieties/"+v.ID.String(), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `value="9"`) {
		t.Errorf("422 should preserve the entered body value; got:\n%s", rec.Body.String())
	}
	// Nothing was persisted.
	got, _ := varieties.GetCharacteristics(context.Background(), v.ID)
	if !got.IsZero() {
		t.Errorf("a rejected edit should persist nothing; got %+v", got)
	}
}

func TestVariety_PutUnknownIs404(t *testing.T) {
	srv, _ := newVarietyTestServer(t)
	form := url.Values{"body": {"3"}, "tannin": {"3"}, "acidity": {"3"}, "sweetness": {"3"}, "alcohol": {"3"}}
	req := httptest.NewRequest(http.MethodPut, "/varieties/nope", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestVariety_GetEditFragmentIsBareForm(t *testing.T) {
	srv, varieties := newVarietyTestServer(t)
	v, _ := domain.NewVariety("Shiraz")
	varieties.Save(v)
	c, _ := domain.NewCharacteristics(5, 4, 2, 1, 5, nil, domain.ProvenanceDefault)
	_ = varieties.SetCharacteristics(context.Background(), v.ID, c)

	req := httptest.NewRequest(http.MethodGet, "/varieties/"+v.ID.String()+"/edit", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if strings.Contains(strings.ToLower(body), "<!doctype html>") {
		t.Errorf("edit fragment should be bare markup, not a full page; got:\n%s", body)
	}
	if !strings.Contains(body, "characteristics-form") {
		t.Errorf("edit fragment should be the form; got:\n%s", body)
	}
}
