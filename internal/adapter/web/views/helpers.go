package views

import "strings"

// DrinkerOption is the switcher's view of a Drinker.
type DrinkerOption struct {
	ID     string
	Name   string
	Active bool
}

// DrinkerSwitcherModel is the view model for the switcher region: the Drinkers
// to choose between (one Active), the name entered in the add form (preserved
// across a failed submit), and a field-to-message error map (empty on first
// paint; the empty-string key carries a form-level banner).
type DrinkerSwitcherModel struct {
	Drinkers []DrinkerOption
	Name     string
	Errors   map[string]string
}

func (m DrinkerSwitcherModel) err(field string) string { return m.Errors[field] }

// activeName is the name of the active Drinker, used to prefill the rename box.
func (m DrinkerSwitcherModel) activeName() string {
	for _, d := range m.Drinkers {
		if d.Active {
			return d.Name
		}
	}
	return ""
}

// activeID is the ID of the active Drinker, the target of the rename form.
func (m DrinkerSwitcherModel) activeID() string {
	for _, d := range m.Drinkers {
		if d.Active {
			return d.ID
		}
	}
	return ""
}

// WineOption is the log form's view of a Wine.
type WineOption struct {
	ID    string
	Label string
}

// CompanionOption is the log form's view of an existing Companion in the active
// Drinker's personal zone, offered for selection.
type CompanionOption struct {
	ID   string
	Name string
}

// LogFormModel is the view model the log-a-Tasting form renders: the Wine
// options to choose from, the values the Drinker entered (preserved across a
// failed submit), and a field-to-message error map. Errors is empty on first
// paint. The empty-string key carries a form-level (non-field) banner.
type LogFormModel struct {
	Wines         []WineOption
	Companions    []CompanionOption
	WineID        string
	Vintage       string
	Rating        string
	Note          string
	CompanionIDs  []string // existing Companions selected (preserved on failed submit)
	NewCompanions string   // free-text new names entered (comma/newline separated)
	Errors        map[string]string
}

func (m LogFormModel) err(field string) string { return m.Errors[field] }

// companionSelected reports whether an existing Companion was chosen, so a
// failed submit re-checks the boxes the Drinker had ticked.
func (m LogFormModel) companionSelected(id string) bool {
	for _, s := range m.CompanionIDs {
		if s == id {
			return true
		}
	}
	return false
}

// VarietyOption is the composition form's view of a pickable Variety.
type VarietyOption struct {
	ID   string
	Name string
}

// CompositionRow is one editable row in the composition form: the chosen
// Variety (empty when the Drinker hasn't picked one yet) and the proportion
// typed. Entered values are preserved across a failed submit.
type CompositionRow struct {
	VarietyID  string
	Proportion string
}

// CompositionFormModel is the view model the edit-Composition form renders: the
// Wine being edited, the Varieties to pick from, the rows entered (preserved
// across a failed submit), and a field-to-message error map. Errors is empty on
// first paint; the empty-string key carries a form-level banner.
type CompositionFormModel struct {
	WineID    string
	WineLabel string
	// Style is the Wine's label-style, if any. When set, the form offers a
	// "fill from {Style} default" affordance that prefills the conventional grapes.
	Style     string
	Varieties []VarietyOption
	Rows      []CompositionRow
	Errors    map[string]string
}

func (m CompositionFormModel) err(field string) string { return m.Errors[field] }

// VarietyCharacteristicsFormModel is the view model the edit-characteristics
// form renders: the Variety being edited, the five scalar axes and the
// flavour-note tags as entered (preserved across a failed submit), and a
// field-to-message error map. Axes are strings so a failed submit re-renders
// exactly what the Drinker typed; Notes is the comma-separated tag text. Errors
// is empty on first paint; the empty-string key carries a form-level banner.
type VarietyCharacteristicsFormModel struct {
	VarietyID                                 string
	VarietyName                               string
	Body, Tannin, Acidity, Sweetness, Alcohol string
	Notes                                     string
	Errors                                    map[string]string
}

func (m VarietyCharacteristicsFormModel) err(field string) string { return m.Errors[field] }

// axisScale renders a 1..5 axis value as filled/empty pips, the same visual
// vocabulary as ratingStars, so the rubric reads at a glance.
func axisScale(v int) string {
	if v < 0 {
		v = 0
	}
	if v > 5 {
		v = 5
	}
	return strings.Repeat("●", v) + strings.Repeat("○", 5-v)
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
