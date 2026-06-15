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

func TestLayoutSwaps422Responses(t *testing.T) {
	html := renderLayout(t)
	if !strings.Contains(html, "htmx.config.responseHandling") {
		t.Errorf("Layout should configure htmx.config.responseHandling; got:\n%s", html)
	}
	if !strings.Contains(html, "422") {
		t.Errorf("Layout responseHandling should add a rule for 422; got:\n%s", html)
	}
}
