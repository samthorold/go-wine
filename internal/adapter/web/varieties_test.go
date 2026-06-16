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
)

func TestVarieties_GetListsSeededVarieties(t *testing.T) {
	drinkers := memory.NewDrinkerRepo()
	wines := memory.NewWineRepo()
	tastings := memory.NewTastingRepo()
	varieties := memory.NewVarietyRepo()
	companions := memory.NewCompanionRepo()

	d, _ := domain.NewDrinker("Sam")
	_ = drinkers.Save(context.Background(), d)
	v, _ := domain.NewVariety("Shiraz")
	varieties.Save(v)

	logH := app.NewLogTastingHandler(drinkers, wines, tastings)
	listH := app.NewListTastingsHandler(wines, tastings, companions)
	listV := app.NewListVarietiesHandler(varieties)
	listW := app.NewListWinesHandler(wines)
	getW := app.NewGetWineHandler(wines, varieties)
	editC := app.NewEditCompositionHandler(wines, varieties)
	createD := app.NewCreateDrinkerHandler(drinkers)
	renameD := app.NewRenameDrinkerHandler(drinkers)
	srv := web.NewServer(drinkers, wines, varieties, companions, logH, listH, listV, listW, getW, editC, createD, renameD)

	req := httptest.NewRequest(http.MethodGet, "/varieties", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Shiraz") {
		t.Errorf("GET /varieties should list the seeded Variety; got:\n%s", body)
	}
}
