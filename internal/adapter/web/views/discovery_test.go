package views

import (
	"strings"
	"testing"

	"go-wine/internal/app"
)

func TestDiscoveryPage_ListsRecommendationsWithExplanation(t *testing.T) {
	recs := []app.RecommendationView{
		{VarietyID: "v1", Name: "Aglianico", Because: []string{"Nebbiolo", "Sangiovese"}},
		{VarietyID: "v2", Name: "Grüner Veltliner", Because: []string{"Riesling"}},
	}
	html := render(t, DiscoveryPage(nil, recs))

	for _, want := range []string{"Aglianico", "Nebbiolo", "Sangiovese", "Grüner Veltliner", "Riesling"} {
		if !strings.Contains(html, want) {
			t.Errorf("discovery page should contain %q; got:\n%s", want, html)
		}
	}
	// The recommendation links to the grape's detail page.
	if !strings.Contains(html, "/varieties/v1") {
		t.Errorf("recommendation should link to the variety detail page; got:\n%s", html)
	}
	// It is a page (Layout shell), not a bare fragment.
	if !strings.Contains(strings.ToLower(html), "<!doctype html>") {
		t.Errorf("DiscoveryPage should be a full page wrapped in Layout; got:\n%s", html)
	}
}

func TestDiscoveryPage_EmptyState(t *testing.T) {
	html := render(t, DiscoveryPage(nil, nil))
	low := strings.ToLower(html)
	// An empty profile gets a clear explanatory state, not a crash or blank list.
	if !strings.Contains(low, "log") || !strings.Contains(low, "love") {
		t.Errorf("empty discovery page should explain how to get recommendations; got:\n%s", html)
	}
}
