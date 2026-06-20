package views

import (
	"strings"
	"testing"
)

func TestDrinkersPage_MarksNoNavLinkActive(t *testing.T) {
	// Drinkers has no top-level nav link, so no link is marked active.
	html := render(t, DrinkersPage(nil, DrinkersModel{}))

	if strings.Contains(html, `aria-current="page"`) {
		t.Errorf("the drinkers page (no nav link) should mark no nav link active; got:\n%s", html)
	}
}

func TestDrinkersManagement_ListsRenameFormsAndAddForm(t *testing.T) {
	model := DrinkersModel{Drinkers: []DrinkerOption{
		{ID: "d1", Name: "Sam", Active: true},
		{ID: "d2", Name: "Partner"},
	}}
	html := render(t, DrinkersManagement(model))

	if !strings.Contains(html, `id="drinkers"`) {
		t.Errorf("management region should own #drinkers; got:\n%s", html)
	}
	if !strings.Contains(html, `hx-put="/drinkers/d1"`) || !strings.Contains(html, `hx-put="/drinkers/d2"`) {
		t.Errorf("should render a rename form per Drinker; got:\n%s", html)
	}
	if !strings.Contains(html, `hx-post="/drinkers"`) {
		t.Errorf("should render the add form; got:\n%s", html)
	}
	// Add/rename are admin chrome, not the page's primary action: Pico secondary.
	if !strings.Contains(html, `<button type="submit" class="secondary">Add</button>`) {
		t.Errorf("Add should be a secondary button; got:\n%s", html)
	}
}

func TestDrinkersManagement_RenameErrorShowsOnlyOnItsRow(t *testing.T) {
	model := DrinkersModel{
		Drinkers: []DrinkerOption{
			{ID: "d1", Name: "Sam"},
			{ID: "d2", Name: "Partner"},
		},
		Errors:        map[string]string{"rename": "please enter a name"},
		RenameErrorID: "d2",
	}
	html := render(t, DrinkersManagement(model))

	if strings.Count(html, "please enter a name") != 1 {
		t.Errorf("rename error should appear exactly once (on its row); got:\n%s", html)
	}
}
