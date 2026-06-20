package views

import (
	"context"
	"strconv"
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

func TestLogForm_RatingIsStarRadioGroup(t *testing.T) {
	// Domain accents are consistent across read and write: a Rating is entered
	// as stars, not a number <select>. The form renders a CSS-only radio group
	// named "rating" with values 1..5. See look-and-feel.md.
	model := LogFormModel{Wines: []WineOption{{ID: "w1", Label: "Penfolds"}}}
	html := render(t, LogForm(model))

	ratingRegion := html
	if i := strings.Index(html, "Rating"); i >= 0 {
		ratingRegion = html[i:]
	}
	if j := strings.Index(ratingRegion, "</fieldset>"); j >= 0 {
		ratingRegion = ratingRegion[:j]
	}
	if strings.Contains(ratingRegion, "<select name=\"rating\"") {
		t.Errorf("rating should not be a <select>; got:\n%s", ratingRegion)
	}
	for r := 1; r <= 5; r++ {
		want := `type="radio" name="rating" value="` + strconv.Itoa(r) + `"`
		if !strings.Contains(html, want) {
			t.Errorf("rating should render a radio for value %d (%q); got:\n%s", r, want, html)
		}
	}
}

func TestLogForm_PreservesChosenRatingOn422(t *testing.T) {
	// On a 422 re-render the chosen rating is preserved: the matching radio is
	// checked, mirroring the old selected?= behaviour.
	model := LogFormModel{
		Wines:  []WineOption{{ID: "w1", Label: "Penfolds"}},
		Rating: "3",
	}
	html := render(t, LogForm(model))

	chunks := strings.Split(html, "<input")
	for _, c := range chunks {
		open := c
		if i := strings.Index(open, ">"); i >= 0 {
			open = open[:i]
		}
		if !strings.Contains(open, `name="rating"`) {
			continue
		}
		isThree := strings.Contains(open, `value="3"`)
		isChecked := strings.Contains(open, "checked")
		if isThree && !isChecked {
			t.Errorf("the chosen rating (3) radio should be checked; got:\n%s", open)
		}
		if !isThree && isChecked {
			t.Errorf("only the chosen rating radio should be checked; got:\n%s", open)
		}
	}
}

func TestLogForm_RatingIsRequired(t *testing.T) {
	// HTML5 constraint: a rating must be chosen. Radios carry required.
	model := LogFormModel{Wines: []WineOption{{ID: "w1", Label: "Penfolds"}}}
	html := render(t, LogForm(model))

	found := false
	for _, c := range strings.Split(html, "<input")[1:] {
		open := c
		if i := strings.Index(open, ">"); i >= 0 {
			open = open[:i]
		}
		if strings.Contains(open, `name="rating"`) && strings.Contains(open, "required") {
			found = true
		}
	}
	if !found {
		t.Errorf("a rating radio should carry required; got:\n%s", html)
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

// countFilledButtons counts <button> elements that carry no Pico variant class
// (secondary/outline) — i.e. the filled-accent primaries. The look-and-feel
// rule is exactly one per page.
func countFilledButtons(html string) int {
	n := 0
	for _, frag := range strings.Split(html, "<button")[1:] {
		open := frag
		if i := strings.Index(open, ">"); i >= 0 {
			open = open[:i]
		}
		if !strings.Contains(open, `class="secondary"`) &&
			!strings.Contains(open, `class="outline"`) {
			n++
		}
	}
	return n
}

func TestTastingsPage_HasExactlyOnePrimaryButton(t *testing.T) {
	drinkers := []DrinkerOption{{ID: "d1", Name: "Sam", Active: true}}
	model := LogFormModel{Wines: []WineOption{{ID: "w1", Label: "Penfolds"}}}
	html := render(t, TastingsPage(drinkers, model, nil))

	if !strings.Contains(html, `<button type="submit">Log tasting</button>`) {
		t.Errorf("Log tasting should be the filled-accent primary; got:\n%s", html)
	}
	if got := countFilledButtons(html); got != 1 {
		t.Errorf("tastings page should have exactly one filled-accent button, got %d;\n%s", got, html)
	}
}

func TestLogForm_ConstrainedToReadableMeasure(t *testing.T) {
	// Readable measure over full bleed: the form/content region carries the
	// .measure content-column class so it does not stretch to the full
	// container width. See look-and-feel.md.
	model := LogFormModel{Wines: []WineOption{{ID: "w1", Label: "Penfolds"}}}
	html := render(t, LogForm(model))

	form := html[strings.Index(html, "<form"):]
	if i := strings.Index(form, ">"); i >= 0 {
		form = form[:i]
	}
	if !strings.Contains(form, "measure") {
		t.Errorf("log form should carry the .measure content-column class; got open tag:\n%s", form)
	}
}

func TestTastingsPage_NavStaysFullWidth(t *testing.T) {
	// The chrome/nav must NOT be constrained to the content measure — only the
	// content column is. See look-and-feel.md.
	drinkers := []DrinkerOption{{ID: "d1", Name: "Sam", Active: true}}
	model := LogFormModel{Wines: []WineOption{{ID: "w1", Label: "Penfolds"}}}
	html := render(t, TastingsPage(drinkers, model, nil))

	nav := html[strings.Index(html, "<nav"):]
	if i := strings.Index(nav, ">"); i >= 0 {
		nav = nav[:i]
	}
	if strings.Contains(nav, "measure") {
		t.Errorf("nav/chrome must stay full-width (no .measure); got open tag:\n%s", nav)
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
