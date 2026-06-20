package views

import (
	"strings"
	"testing"

	"go-wine/internal/app"
)

func TestWineDetailPage_HasExactlyOnePrimaryButton(t *testing.T) {
	// Layout carries the demoted Add/Rename chrome and the form carries the
	// outline Fill-from-default helper; the page's one filled-accent primary is
	// Save grapes. See look-and-feel.md.
	drinkers := []DrinkerOption{{ID: "d1", Name: "Sam", Active: true}}
	wine := app.WineDetailView{ID: "w1", Label: "Penfolds — Bin 28 Shiraz", Style: "Shiraz"}
	form := CompositionFormModel{WineID: "w1", Style: "Shiraz"}
	html := render(t, WineDetailPage(drinkers, wine, form, app.WineVerdictView{}))

	if !strings.Contains(html, `<button type="submit">Save grapes</button>`) {
		t.Errorf("Save grapes should be the filled-accent primary; got:\n%s", html)
	}
	if got := countFilledButtons(html); got != 1 {
		t.Errorf("wine detail page should have exactly one filled-accent button, got %d;\n%s", got, html)
	}
}

func TestCompositionForm_SaveIsPrimaryFillIsOutline(t *testing.T) {
	// The Wine detail view exists to edit grapes, so "Save grapes" is the one
	// filled-accent primary. "Fill from … default" is a secondary helper and
	// recedes to Pico's outline variant. See look-and-feel.md.
	model := CompositionFormModel{WineID: "w1", Style: "Shiraz"}
	html := render(t, CompositionForm(model))

	if !strings.Contains(html, `<button type="submit">Save grapes</button>`) {
		t.Errorf("Save grapes should stay the filled-accent primary (no variant class); got:\n%s", html)
	}
	if !strings.Contains(html, `class="outline"`) {
		t.Errorf("Fill-from-default helper should be Pico outline; got:\n%s", html)
	}
	if !strings.Contains(html, "Fill from Shiraz default") {
		t.Errorf("form should render the fill-from-default helper for a styled wine; got:\n%s", html)
	}
}
