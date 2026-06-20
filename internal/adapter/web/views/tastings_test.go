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

func TestLogForm_WineDefaultsToPlaceholderNotARealWine(t *testing.T) {
	// Wine is the one genuinely required field, so a blank form must NOT
	// pre-select a real Wine: the select defaults to a non-submittable
	// "Choose a wine…" placeholder. The placeholder carries an empty value and
	// is disabled so HTML5 `required` blocks submit while it is the selection,
	// and no real Wine option is selected. See issue #40.
	model := LogFormModel{
		Wines: []WineOption{{ID: "w1", Label: "Cloudy Bay — Sauvignon Blanc"}},
	}
	html := render(t, LogForm(model))

	// The select is still required (HTML5 first line of defence).
	if !strings.Contains(html, `<select name="wine_id" required>`) {
		t.Errorf("wine select should be required; got:\n%s", html)
	}

	// A disabled, empty-valued placeholder exists and is the selected option.
	if !strings.Contains(html, `value="" disabled selected`) {
		t.Errorf("wine select should default to a disabled, empty-valued, selected placeholder; got:\n%s", html)
	}
	if !strings.Contains(html, "Choose a wine") {
		t.Errorf("placeholder should read \"Choose a wine…\"; got:\n%s", html)
	}

	// No real Wine option is selected on a blank form.
	for _, frag := range strings.Split(html, "<option")[1:] {
		open := frag
		if i := strings.Index(open, ">"); i >= 0 {
			open = open[:i]
		}
		if strings.Contains(open, `value="w1"`) && strings.Contains(open, "selected") {
			t.Errorf("no real Wine should be selected on a blank form; got:\n%s", open)
		}
	}
}

func TestLogForm_PreservesChosenWineOn422(t *testing.T) {
	// On a 422 re-render the chosen Wine is preserved: its option is selected
	// and the placeholder is not. See issue #40.
	model := LogFormModel{
		Wines:  []WineOption{{ID: "w1", Label: "Cloudy Bay — Sauvignon Blanc"}},
		WineID: "w1",
	}
	html := render(t, LogForm(model))

	for _, frag := range strings.Split(html, "<option")[1:] {
		open := frag
		if i := strings.Index(open, ">"); i >= 0 {
			open = open[:i]
		}
		isReal := strings.Contains(open, `value="w1"`)
		isPlaceholder := strings.Contains(open, `value=""`)
		isSelected := strings.Contains(open, "selected")
		if isReal && !isSelected {
			t.Errorf("the chosen Wine option should be selected on 422; got:\n%s", open)
		}
		if isPlaceholder && isSelected {
			t.Errorf("the placeholder must not be selected when a real Wine is chosen; got:\n%s", open)
		}
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

func TestLogForm_FirstPaintHasNoRatingSelected(t *testing.T) {
	// A fresh tasting starts with NO rating selected so a Drinker cannot save a
	// phantom rating. None of the 1..5 value-bearing radios is checked; a
	// "no rating" option (value="") is checked instead. See issue #39.
	model := LogFormModel{Wines: []WineOption{{ID: "w1", Label: "Penfolds"}}}
	html := render(t, LogForm(model))

	for _, c := range strings.Split(html, "<input")[1:] {
		open := c
		if i := strings.Index(open, ">"); i >= 0 {
			open = open[:i]
		}
		if !strings.Contains(open, `name="rating"`) {
			continue
		}
		valued := !strings.Contains(open, `value=""`)
		checked := strings.Contains(open, "checked")
		if valued && checked {
			t.Errorf("no value-bearing rating radio should be checked on first paint; got:\n%s", open)
		}
	}

	// The "no rating" option exists and is the one checked by default.
	if !strings.Contains(html, `name="rating" value=""`) {
		t.Errorf("form should offer a no-rating option (value=\"\"); got:\n%s", html)
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

func TestLogForm_RatingIsNotHTML5Required(t *testing.T) {
	// A fresh tasting starts unrated and can be submitted as "no rating": the
	// domain command handler is the authority that rejects it (with an inline
	// 422 error), not an HTML5 required constraint — which a checked value=""
	// option would satisfy anyway, making it meaningless. See issue #39.
	model := LogFormModel{Wines: []WineOption{{ID: "w1", Label: "Penfolds"}}}
	html := render(t, LogForm(model))

	for _, c := range strings.Split(html, "<input")[1:] {
		open := c
		if i := strings.Index(open, ">"); i >= 0 {
			open = open[:i]
		}
		if strings.Contains(open, `name="rating"`) && strings.Contains(open, "required") {
			t.Errorf("rating radios must not carry HTML5 required; the domain is the authority; got:\n%s", open)
		}
	}
}

func TestLogForm_OffersClearToNoRating(t *testing.T) {
	// The Drinker can clear a selected rating back to "none": a checked rating
	// is still accompanied by the no-rating option so a click returns to unrated
	// without JS. See issue #39.
	model := LogFormModel{Wines: []WineOption{{ID: "w1", Label: "Penfolds"}}, Rating: "3"}
	html := render(t, LogForm(model))

	if !strings.Contains(html, `name="rating" value=""`) {
		t.Errorf("a rated form should still offer the no-rating option to clear; got:\n%s", html)
	}
	if !strings.Contains(html, "rating-clear") {
		t.Errorf("the no-rating option should carry a Clear affordance; got:\n%s", html)
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

func TestTastingsPage_ContentSitsInReadableMeasure(t *testing.T) {
	// Readable measure over full bleed: the tasting page's content (form + list)
	// sits inside the shared .measure content column so it does not stretch to the
	// full container width. The constraint comes from the shared Layout wrapper —
	// applied once for every page — not from a per-form class, so there is a single
	// source of the column width. See issue #45 / look-and-feel.md.
	drinkers := []DrinkerOption{{ID: "d1", Name: "Sam", Active: true}}
	model := LogFormModel{Wines: []WineOption{{ID: "w1", Label: "Penfolds"}}}
	html := render(t, TastingsPage(drinkers, model, nil))

	// The log form sits inside the shared .measure column...
	col := strings.Index(html, `class="measure"`)
	if col < 0 {
		t.Fatalf("tastings page content should sit in the shared .measure column; got:\n%s", html)
	}
	if !strings.Contains(html[col:], `id="log-form"`) {
		t.Errorf("log form should sit inside the shared .measure column; got:\n%s", html)
	}

	// ...and the form itself no longer carries a redundant per-form measure class,
	// so the column width has a single source and the two cannot drift.
	form := html[strings.Index(html, `id="log-form"`):]
	if i := strings.Index(form, ">"); i >= 0 {
		form = form[:i]
	}
	if strings.Contains(form, "measure") {
		t.Errorf("log form must not redundantly carry .measure (shared column owns it); got open tag:\n%s", form)
	}
}

func TestLogForm_VintageReadsAsHintNotValue(t *testing.T) {
	// The Vintage placeholder must not read like a pre-filled value (a bare year
	// like "2019"); an empty field should be unambiguous. A <small> helper makes
	// the cue clearly guidance. See issue #32 / look-and-feel.md.
	model := LogFormModel{Wines: []WineOption{{ID: "w1", Label: "Penfolds"}}}
	html := render(t, LogForm(model))

	if strings.Contains(html, `placeholder="2019"`) {
		t.Errorf("Vintage placeholder must not read as a real value (\"2019\"); got:\n%s", html)
	}
	if !strings.Contains(html, "<small>") {
		t.Errorf("Vintage field should carry a <small> hint so an empty field is unambiguous; got:\n%s", html)
	}
}

func TestTastingsPage_MarksTastingsNavActive(t *testing.T) {
	// Active-nav state: the nav link for the current page carries
	// aria-current="page" so Pico styles it, and no other link does.
	drinkers := []DrinkerOption{{ID: "d1", Name: "Sam", Active: true}}
	model := LogFormModel{Wines: []WineOption{{ID: "w1", Label: "Penfolds"}}}
	html := render(t, TastingsPage(drinkers, model, nil))

	if !strings.Contains(html, `<a href="/tastings" aria-current="page">Tastings</a>`) {
		t.Errorf("Tastings nav link should carry aria-current=\"page\"; got:\n%s", html)
	}
	if strings.Contains(html, `<a href="/wines" aria-current`) ||
		strings.Contains(html, `<a href="/varieties" aria-current`) ||
		strings.Contains(html, `<a href="/discovery" aria-current`) {
		t.Errorf("only the current page's nav link should be active; got:\n%s", html)
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

func TestCompanions_OwnsRegionAndOffersExistingPlusAddControl(t *testing.T) {
	// The Companions region is its own component owning #companions, so first
	// paint and the add-Companion re-render render from one source and cannot
	// drift. It offers existing Companions to tick, plus a clearly-styled
	// "+ Add companion" control that posts to the add sub-resource with an
	// explicit target/swap (never relying on htmx defaults). See issue #44.
	model := LogFormModel{
		Companions: []CompanionOption{{ID: "c1", Name: "Alex"}, {ID: "c2", Name: "Jo"}},
	}
	html := render(t, Companions(model))

	if !strings.Contains(html, `id="companions"`) {
		t.Errorf("Companions should own the #companions region; got:\n%s", html)
	}
	if !strings.Contains(html, "Alex") || !strings.Contains(html, "Jo") {
		t.Errorf("region should offer existing companions to pick; got:\n%s", html)
	}
	if !strings.Contains(html, `name="companion_id"`) {
		t.Errorf("region should let you pick existing companions by id; got:\n%s", html)
	}
	// The add control is an explicit-swap POST onto the region itself.
	if !strings.Contains(html, `hx-post="/tastings/companions"`) {
		t.Errorf("add control should post to the add sub-resource; got:\n%s", html)
	}
	if !strings.Contains(html, `hx-target="#companions"`) {
		t.Errorf("add control should target #companions explicitly; got:\n%s", html)
	}
	if !strings.Contains(html, `hx-swap="outerHTML"`) {
		t.Errorf("add control should set hx-swap explicitly; got:\n%s", html)
	}
	if !strings.Contains(html, `name="new_companion"`) {
		t.Errorf("add control should carry a name field for the new companion; got:\n%s", html)
	}
	// The control reads as a control (a button), not bare "Add new" text.
	if !strings.Contains(html, "Add companion") {
		t.Errorf("add control should be labelled; got:\n%s", html)
	}
}

// LogForm embeds the Companions region so the picker is present on first paint.
func TestLogForm_EmbedsCompanionsRegion(t *testing.T) {
	model := LogFormModel{
		Wines:      []WineOption{{ID: "w1", Label: "Penfolds — Bin 28 Shiraz"}},
		Companions: []CompanionOption{{ID: "c1", Name: "Alex"}},
	}
	html := render(t, LogForm(model))

	if !strings.Contains(html, `id="companions"`) {
		t.Errorf("log form should embed the #companions region; got:\n%s", html)
	}
	if !strings.Contains(html, "Alex") {
		t.Errorf("embedded region should render existing companions; got:\n%s", html)
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
