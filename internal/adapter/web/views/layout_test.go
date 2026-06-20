package views

import (
	"context"
	"strings"
	"testing"
)

// renderLayout renders the Layout shell (with no children) to a string.
func renderLayout(t *testing.T) string {
	t.Helper()
	var sb strings.Builder
	if err := Layout("Tastings", nil).Render(context.Background(), &sb); err != nil {
		t.Fatalf("rendering Layout: %v", err)
	}
	return sb.String()
}

func TestLayoutBoostsNavigation(t *testing.T) {
	html := renderLayout(t)
	if !strings.Contains(html, `hx-boost="true"`) {
		t.Errorf("Layout body should be boosted with hx-boost=\"true\"; got:\n%s", html)
	}
}

func TestDrinkerSwitcherPostsToSwitch(t *testing.T) {
	var sb strings.Builder
	opts := []DrinkerOption{{ID: "d1", Name: "Sam", Active: true}}
	if err := Layout("Tastings", opts).Render(context.Background(), &sb); err != nil {
		t.Fatalf("rendering Layout: %v", err)
	}
	html := sb.String()
	if !strings.Contains(html, `method="post"`) {
		t.Errorf("switcher form should POST (no safe-method mutation); got:\n%s", html)
	}
	if strings.Contains(html, `method="get"`) {
		t.Errorf("switcher form should not use method=get; got:\n%s", html)
	}
}

func TestDrinkerSwitcherButtonsAreSecondary(t *testing.T) {
	// Add/Rename are chrome/admin actions, not the view's primary action.
	// They must recede to Pico's secondary variant so the page keeps a single
	// filled-accent primary (e.g. Log tasting). See look-and-feel.md.
	opts := []DrinkerOption{{ID: "d1", Name: "Sam", Active: true}}
	html := render(t, DrinkerSwitcher(DrinkerSwitcherModel{Drinkers: opts}))

	if !strings.Contains(html, `Add`) || !strings.Contains(html, `Rename`) {
		t.Fatalf("switcher should render Add and Rename; got:\n%s", html)
	}
	for _, want := range []string{
		`<button type="submit" class="secondary">Add</button>`,
		`<button type="submit" class="secondary">Rename</button>`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("chrome button should be Pico secondary: want %q; got:\n%s", want, html)
		}
	}
}

func TestLayoutSwaps422Responses(t *testing.T) {
	html := renderLayout(t)
	if !strings.Contains(html, "htmx.config.responseHandling") {
		t.Errorf("Layout should configure htmx.config.responseHandling; got:\n%s", html)
	}
	if !strings.Contains(html, "422") {
		t.Errorf("Layout responseHandling should add a rule for 422; got:\n%s", html)
	}
}
