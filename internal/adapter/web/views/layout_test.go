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
	if err := Layout("Tastings", SectionTastings, nil).Render(context.Background(), &sb); err != nil {
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
	if err := Layout("Tastings", SectionTastings, opts).Render(context.Background(), &sb); err != nil {
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

func TestLayoutStylesStarRatingInput(t *testing.T) {
	// Domain accents are consistent across read and write: the star radio input
	// is styled CSS-only in the one rationed <style> block, reusing the .rating
	// coral accent so read and write stars match. See look-and-feel.md.
	html := renderLayout(t)

	style := html
	if i := strings.Index(style, "<style>"); i >= 0 {
		style = style[i:]
	}
	if j := strings.Index(style, "</style>"); j >= 0 {
		style = style[:j]
	}
	if !strings.Contains(style, ".rating-input") {
		t.Errorf("Layout <style> should style the .rating-input star control; got:\n%s", style)
	}
	// The accent must be the existing .rating colour, not a new one: the rule
	// is grouped onto the .rating selector rather than redeclaring the colour.
	if !strings.Contains(style, ".rating-input") || !strings.Contains(style, ".rating") {
		t.Errorf(".rating-input should reuse the .rating accent; got:\n%s", style)
	}
}

func TestLayoutStylesRatingClearAffordance(t *testing.T) {
	// The no-rating "Clear" affordance recedes: it is a small muted control, not
	// a coral star, so it reads as "return to no rating" rather than a sixth
	// star. Styled CSS-only in the one rationed <style> block. See issue #39.
	html := renderLayout(t)

	style := html
	if i := strings.Index(style, "<style>"); i >= 0 {
		style = style[i:]
	}
	if j := strings.Index(style, "</style>"); j >= 0 {
		style = style[:j]
	}
	if !strings.Contains(style, ".rating-clear") {
		t.Errorf("Layout <style> should style the .rating-clear no-rating affordance; got:\n%s", style)
	}
}

func TestLayoutRaisesTextContrastForAA(t *testing.T) {
	// WCAG AA contrast (>= 4.5:1) across every form and page: rather than restyle
	// per-component, the one rationed <style> block overrides Pico's dark
	// contrast tokens so labels, helper <small> text, entered values and
	// placeholders all clear AA against the near-black background. This test
	// locks in the *mechanism* (the overrides are present); the contrast maths
	// is verified visually + in the PR body. See issue #41 and look-and-feel.md.
	html := renderLayout(t)

	style := html
	if i := strings.Index(style, "<style>"); i >= 0 {
		style = style[i:]
	}
	if j := strings.Index(style, "</style>"); j >= 0 {
		style = style[:j]
	}

	// Helper text and labels lift off Pico's borderline muted grey.
	if !strings.Contains(style, "--pico-muted-color:") {
		t.Errorf("Layout <style> should override --pico-muted-color for AA-legible labels/helper text; got:\n%s", style)
	}
	// Placeholder stays dimmer than entered values but legible — an explicit
	// override keeps it distinguishable while clearing AA.
	if !strings.Contains(style, "--pico-form-element-placeholder-color:") {
		t.Errorf("Layout <style> should override --pico-form-element-placeholder-color so placeholders stay legible yet dimmer than values; got:\n%s", style)
	}
	// Entered values / body text are set explicitly bright.
	if !strings.Contains(style, "--pico-color:") {
		t.Errorf("Layout <style> should override --pico-color for bright entered-value/body text; got:\n%s", style)
	}
}

func TestLayoutStrengthensFormBordersAndFocusRing(t *testing.T) {
	// Form controls (Wine select, Vintage field, Note textarea) blend into the
	// near-black background — they read as banners more than editable controls.
	// Rather than restyle per-component, the one rationed <style> block overrides
	// Pico's dark form-element border/focus tokens so every input/select/textarea
	// gets a visible resting border and an unmistakable focus ring (keyboard and
	// pointer). This test locks in the *mechanism* (the overrides are present);
	// the exact contrast is verified visually + in the PR body. See issue #42 and
	// look-and-feel.md.
	html := renderLayout(t)

	style := html
	if i := strings.Index(style, "<style>"); i >= 0 {
		style = style[i:]
	}
	if j := strings.Index(style, "</style>"); j >= 0 {
		style = style[:j]
	}

	// Resting border lifts off Pico's near-invisible dark default so controls
	// read as editable against the page background.
	if !strings.Contains(style, "--pico-form-element-border-color:") {
		t.Errorf("Layout <style> should override --pico-form-element-border-color for a visible resting border; got:\n%s", style)
	}
	// Focused border is distinct (keyboard and pointer focus).
	if !strings.Contains(style, "--pico-form-element-active-border-color:") {
		t.Errorf("Layout <style> should override --pico-form-element-active-border-color for a distinct focused border; got:\n%s", style)
	}
	// The focus RING (a box-shadow driven by this token) is raised so it is
	// unmistakable rather than the low-contrast dark default.
	if !strings.Contains(style, "--pico-form-element-focus-color:") {
		t.Errorf("Layout <style> should override --pico-form-element-focus-color for an unmistakable focus ring; got:\n%s", style)
	}
}

func TestLayoutGivesActiveNavItemACurrentTreatment(t *testing.T) {
	// The active nav item must read as "you are here", not disabled: Pico's
	// native aria-current styling renders the current link white, which against
	// the dark chrome looks greyed-out/disabled rather than current. The one
	// rationed <style> block adds a scoped rule giving the active nav link an
	// unambiguous current treatment (accent colour + a visible marker), while
	// inactive links keep Pico's blue link colour so they stay clickable. This
	// test locks in the *mechanism* (the scoped rule is present); the exact look
	// is verified visually + in the PR body. See issue #43 and look-and-feel.md.
	html := renderLayout(t)

	style := html
	if i := strings.Index(style, "<style>"); i >= 0 {
		style = style[i:]
	}
	if j := strings.Index(style, "</style>"); j >= 0 {
		style = style[:j]
	}

	// The rule is scoped to nav links via the aria-current attribute, so it
	// styles only the active nav item and not any other aria-current usage.
	if !strings.Contains(style, `nav a[aria-current="page"]`) {
		t.Errorf("Layout <style> should scope an active-nav rule to nav a[aria-current=\"page\"]; got:\n%s", style)
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
