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
	html := render(t, VarietiesPage(nil, varieties))

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
	html := render(t, VarietiesPage(nil, nil))
	if !strings.Contains(strings.ToLower(html), "no varieties") {
		t.Errorf("empty page should render an empty-state message; got:\n%s", html)
	}
}
