package views

import "strings"

// DrinkerOption is the switcher's view of a Drinker.
type DrinkerOption struct {
	ID     string
	Name   string
	Active bool
}

// WineOption is the log form's view of a Wine.
type WineOption struct {
	ID    string
	Label string
}

// LogFormModel is the view model the log-a-Tasting form renders: the Wine
// options to choose from, the values the Drinker entered (preserved across a
// failed submit), and a field-to-message error map. Errors is empty on first
// paint. The empty-string key carries a form-level (non-field) banner.
type LogFormModel struct {
	Wines   []WineOption
	WineID  string
	Vintage string
	Rating  string
	Note    string
	Errors  map[string]string
}

func (m LogFormModel) err(field string) string { return m.Errors[field] }

func ratingStars(r int) string {
	if r < 0 {
		r = 0
	}
	if r > 5 {
		r = 5
	}
	return strings.Repeat("★", r) + strings.Repeat("☆", 5-r)
}
