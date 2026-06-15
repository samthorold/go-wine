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

func ratingStars(r int) string {
	if r < 0 {
		r = 0
	}
	if r > 5 {
		r = 5
	}
	return strings.Repeat("★", r) + strings.Repeat("☆", 5-r)
}
