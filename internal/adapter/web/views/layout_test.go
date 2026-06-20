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

func TestDrinkerSwitcherIsSwitchOnly(t *testing.T) {
	// Density follows frequency: the nav switcher keeps only the everyday
	// control — the active-Drinker select — plus a link to the management page.
	// Managing the set of Drinkers (add/rename) is a noun, not a chrome widget,
	// so it must NOT appear in the switcher. See look-and-feel.md.
	opts := []DrinkerOption{{ID: "d1", Name: "Sam", Active: true}}
	html := render(t, DrinkerSwitcher(DrinkerSwitcherModel{Drinkers: opts}))

	if strings.Contains(html, "Add") || strings.Contains(html, "Rename") {
		t.Errorf("switcher must not host add/rename admin controls; got:\n%s", html)
	}
	if strings.Contains(html, `hx-post="/drinkers"`) || strings.Contains(html, "hx-put") {
		t.Errorf("switcher must not host the add/rename mutations; got:\n%s", html)
	}
	if !strings.Contains(html, `href="/drinkers"`) {
		t.Errorf("switcher should link to the /drinkers management page; got:\n%s", html)
	}
}

func TestLayoutDefinesReadableMeasure(t *testing.T) {
	// Readable measure over full bleed: the shell (<main class="container">)
	// stays wide, but forms/content are constrained to a comfortable column.
	// That column class is defined in the one rationed <style> block.
	// See look-and-feel.md.
	html := renderLayout(t)
	if !strings.Contains(html, ".measure") {
		t.Errorf("Layout <style> should define a .measure content-column class; got:\n%s", html)
	}
	if !strings.Contains(html, "max-width") {
		t.Errorf(".measure should constrain width via max-width; got:\n%s", html)
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
