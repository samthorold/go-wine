package views

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/a-h/templ"

	"go-wine/internal/app"
)

func render(t *testing.T, c templ.Component) string {
	t.Helper()
	var sb strings.Builder
	if err := c.Render(context.Background(), &sb); err != nil {
		t.Fatalf("render: %v", err)
	}
	return sb.String()
}

func TestTastingList_Populated(t *testing.T) {
	tastings := []app.TastingView{
		{WineLabel: "Penfolds — Bin 28 Shiraz", Rating: 4, DrunkOn: time.Now()},
	}
	html := render(t, TastingList(tastings))

	if !strings.Contains(html, `id="tastings"`) {
		t.Errorf("TastingList should own #tastings; got:\n%s", html)
	}
	if !strings.Contains(html, "Penfolds — Bin 28 Shiraz") {
		t.Errorf("populated list should render the tasting; got:\n%s", html)
	}
}

func TestTastingList_EmptyState(t *testing.T) {
	html := render(t, TastingList(nil))

	if !strings.Contains(html, `id="tastings"`) {
		t.Errorf("TastingList should own #tastings even when empty; got:\n%s", html)
	}
	if strings.Contains(html, "<article>") {
		t.Errorf("empty list should render no tasting rows; got:\n%s", html)
	}
	if !strings.Contains(strings.ToLower(html), "no tastings") {
		t.Errorf("empty list should render an empty-state message; got:\n%s", html)
	}
}

func TestLogForm_FirstPaint(t *testing.T) {
	model := LogFormModel{
		Wines: []WineOption{{ID: "w1", Label: "Penfolds — Bin 28 Shiraz"}},
	}
	html := render(t, LogForm(model))

	if !strings.Contains(html, `id="log-form"`) {
		t.Errorf("form should own id=log-form; got:\n%s", html)
	}
	if !strings.Contains(html, `hx-target="#log-form"`) {
		t.Errorf("form should target #log-form; got:\n%s", html)
	}
	if !strings.Contains(html, `hx-swap="outerHTML"`) {
		t.Errorf("form should swap outerHTML; got:\n%s", html)
	}
	if !strings.Contains(html, "Penfolds — Bin 28 Shiraz") {
		t.Errorf("form should render wine options; got:\n%s", html)
	}
}

func TestLogForm_PreservesValuesAndShowsFieldError(t *testing.T) {
	model := LogFormModel{
		Wines:  []WineOption{{ID: "w1", Label: "Penfolds — Bin 28 Shiraz"}},
		WineID: "w1",
		Note:   "lamb stew",
		Errors: map[string]string{"rating": "rating must be between 1 and 5"},
	}
	html := render(t, LogForm(model))

	if !strings.Contains(html, "rating must be between 1 and 5") {
		t.Errorf("form should render the inline rating error; got:\n%s", html)
	}
	if !strings.Contains(html, "lamb stew") {
		t.Errorf("form should preserve the entered note; got:\n%s", html)
	}
}

func TestLogForm_RendersFormLevelBanner(t *testing.T) {
	model := LogFormModel{
		Errors: map[string]string{"": "something went wrong"},
	}
	html := render(t, LogForm(model))

	if !strings.Contains(html, "something went wrong") {
		t.Errorf("form should render a form-level banner for non-field errors; got:\n%s", html)
	}
}

func TestTastingRow_ShowsCompanions(t *testing.T) {
	view := app.TastingView{
		WineLabel:  "Penfolds — Bin 28 Shiraz",
		Rating:     4,
		Companions: []string{"Alex", "Jo"},
		DrunkOn:    time.Now(),
	}
	html := render(t, TastingRow(view))

	if !strings.Contains(html, "Alex") || !strings.Contains(html, "Jo") {
		t.Errorf("row should show the companions; got:\n%s", html)
	}
}

func TestLogForm_OffersExistingCompanionsAndNewInput(t *testing.T) {
	model := LogFormModel{
		Wines:      []WineOption{{ID: "w1", Label: "Penfolds — Bin 28 Shiraz"}},
		Companions: []CompanionOption{{ID: "c1", Name: "Alex"}, {ID: "c2", Name: "Jo"}},
	}
	html := render(t, LogForm(model))

	if !strings.Contains(html, "Alex") || !strings.Contains(html, "Jo") {
		t.Errorf("form should offer existing companions to pick; got:\n%s", html)
	}
	if !strings.Contains(html, `name="companion_id"`) {
		t.Errorf("form should let you pick existing companions by id; got:\n%s", html)
	}
	if !strings.Contains(html, `name="new_companions"`) {
		t.Errorf("form should let you add new companion names; got:\n%s", html)
	}
}

func TestTastingList_OOB(t *testing.T) {
	html := render(t, TastingListOOB(nil))
	if !strings.Contains(html, `hx-swap-oob="true"`) {
		t.Errorf("OOB list should carry hx-swap-oob=true; got:\n%s", html)
	}
	if !strings.Contains(html, `id="tastings"`) {
		t.Errorf("OOB list should still own #tastings; got:\n%s", html)
	}
}
