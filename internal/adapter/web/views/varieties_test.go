package views

import (
	"strings"
	"testing"

	"go-wine/internal/app"
)

func TestVarietiesPage_ListsVarieties(t *testing.T) {
	varieties := []app.VarietyView{
		{ID: "v1", Name: "Shiraz"},
		{ID: "v2", Name: "Chardonnay"},
	}
	html := render(t, VarietiesPage(nil, varieties, app.TasteProfileView{}))

	if !strings.Contains(html, "Shiraz") {
		t.Errorf("page should list Shiraz; got:\n%s", html)
	}
	if !strings.Contains(html, "Chardonnay") {
		t.Errorf("page should list Chardonnay; got:\n%s", html)
	}
	// It is a page (Layout shell), not a bare fragment.
	if !strings.Contains(strings.ToLower(html), "<!doctype html>") {
		t.Errorf("VarietiesPage should be a full page wrapped in Layout; got:\n%s", html)
	}
}

func TestVarietiesPage_EmptyState(t *testing.T) {
	html := render(t, VarietiesPage(nil, nil, app.TasteProfileView{}))
	if !strings.Contains(strings.ToLower(html), "no varieties") {
		t.Errorf("empty page should render an empty-state message; got:\n%s", html)
	}
}

func TestVarietiesPage_LinksToDetail(t *testing.T) {
	varieties := []app.VarietyView{{ID: "v1", Name: "Shiraz"}}
	html := render(t, VarietiesPage(nil, varieties, app.TasteProfileView{}))
	if !strings.Contains(html, `/varieties/v1`) {
		t.Errorf("variety should link to its detail page; got:\n%s", html)
	}
}

func TestVarietyDetailPage_ShowsCharacteristics(t *testing.T) {
	view := app.VarietyDetailView{
		ID: "v1", Name: "Shiraz",
		Body: 5, Tannin: 4, Acidity: 2, Sweetness: 1, Alcohol: 5,
		Notes: []string{"blackberry", "pepper"}, HasCharacteristics: true,
	}
	form := VarietyCharacteristicsFormModel{VarietyID: "v1", Body: "5", Tannin: "4", Acidity: "2", Sweetness: "1", Alcohol: "5", Notes: "blackberry, pepper"}
	html := render(t, VarietyDetailPage(nil, view, form, app.VarietyPreferenceView{}))

	for _, want := range []string{"Shiraz", "Body", "Tannin", "blackberry", "pepper"} {
		if !strings.Contains(html, want) {
			t.Errorf("detail page should contain %q; got:\n%s", want, html)
		}
	}
	if !strings.Contains(strings.ToLower(html), "<!doctype html>") {
		t.Errorf("VarietyDetailPage should be a full page; got:\n%s", html)
	}
}

func TestVarietyDetailPage_ConfirmedBadge(t *testing.T) {
	confirmed := app.VarietyDetailView{ID: "v1", Name: "Shiraz", Body: 5, Tannin: 4, Acidity: 2, Sweetness: 1, Alcohol: 5, HasCharacteristics: true, Confirmed: true}
	html := render(t, VarietyDetailPage(nil, confirmed, VarietyCharacteristicsFormModel{VarietyID: "v1"}, app.VarietyPreferenceView{}))
	if !strings.Contains(strings.ToLower(html), "confirmed") {
		t.Errorf("a confirmed bundle should show its provenance; got:\n%s", html)
	}

	def := app.VarietyDetailView{ID: "v1", Name: "Shiraz", Body: 5, Tannin: 4, Acidity: 2, Sweetness: 1, Alcohol: 5, HasCharacteristics: true, Confirmed: false}
	html = render(t, VarietyDetailPage(nil, def, VarietyCharacteristicsFormModel{VarietyID: "v1"}, app.VarietyPreferenceView{}))
	if !strings.Contains(strings.ToLower(html), "default") {
		t.Errorf("a default bundle should show its provenance; got:\n%s", html)
	}
}

func TestVarietyCharacteristicsForm_PreservesValuesAndErrors(t *testing.T) {
	model := VarietyCharacteristicsFormModel{
		VarietyID: "v1", Body: "9", Tannin: "4", Acidity: "2", Sweetness: "1", Alcohol: "5",
		Notes:  "blackberry",
		Errors: map[string]string{"": "axes must be between 1 and 5"},
	}
	html := render(t, VarietyCharacteristicsForm(model))
	if !strings.Contains(html, `value="9"`) {
		t.Errorf("form should preserve the entered body value; got:\n%s", html)
	}
	if !strings.Contains(html, "axes must be between 1 and 5") {
		t.Errorf("form should render the error banner; got:\n%s", html)
	}
	if !strings.Contains(html, `hx-put="/varieties/v1"`) {
		t.Errorf("form should PUT to the variety resource; got:\n%s", html)
	}
}
