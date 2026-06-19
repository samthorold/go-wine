package domain

// Variety characteristics are intrinsic, conventional reference data about a
// grape — the scalar axes (body, tannin, acidity, sweetness, alcohol) plus
// typical flavour-note tags. They are the same regardless of who drinks it and
// cover grapes the Drinker has never tried, which is what lets Discovery reach
// beyond their own history. Seeded neutrally against one coherent rubric,
// Provenance-tagged, and only rarely hand-edited.
//
// The rubric: every axis is an integer on a fixed 1..5 scale, scored against a
// single shared yardstick so that proximity in characteristics space is
// meaningful (1 = lowest/lightest, 5 = highest/fullest for that axis). Sharing
// one scale across every Variety is what makes Discovery's distances comparable
// rather than noise.

// Axis is one scalar characteristic of a grape on the fixed 1..5 rubric. Like
// Rating it is a value object that cannot exist outside its scale.
type Axis int

// NewAxis constructs an Axis, rejecting anything outside 1..5.
func NewAxis(v int) (Axis, error) {
	if v < 1 || v > 5 {
		return 0, ErrInvalidCharacteristics
	}
	return Axis(v), nil
}

func (a Axis) Int() int { return int(a) }

// Provenance is the origin of a seeded value — minimal and binary for now: a
// neutral conventional default, or a value confirmed/edited by the Drinker.
// Confirmed values must survive a re-seed (see MergeCharacteristics).
type Provenance int

const (
	// ProvenanceDefault is the neutral conventional seed — a guess, overwritable
	// by a re-seed.
	ProvenanceDefault Provenance = iota
	// ProvenanceConfirmed is a value the Drinker has vetted by editing it; a
	// re-seed never clobbers it.
	ProvenanceConfirmed
)

// Characteristics is a Variety's intrinsic characteristics bundle: the five
// scalar axes, the flavour-note tags, and a single binary Provenance for the
// whole bundle. It is a value object on the Variety aggregate — no identity of
// its own — owned by the VarietyRepository alongside the Variety's name.
//
// Provenance is carried per-bundle rather than per-axis: the model is binary
// (default vs confirmed-by-me) and a hand-edit vets the grape's profile as a
// whole, so one flag is sufficient and keeps the no-clobber merge simple.
type Characteristics struct {
	Body      Axis
	Tannin    Axis
	Acidity   Axis
	Sweetness Axis
	Alcohol   Axis
	// Notes are typical flavour-note tags (e.g. "cherry", "leather"). Categorical
	// set membership, combined as overlap by Discovery — never a magnitude.
	Notes      []string
	Provenance Provenance
}

// NewCharacteristics constructs a validated Characteristics bundle. Every axis
// must lie on the 1..5 rubric; notes are taken as given (an empty set is fine).
func NewCharacteristics(body, tannin, acidity, sweetness, alcohol int, notes []string, p Provenance) (Characteristics, error) {
	axes := [5]int{body, tannin, acidity, sweetness, alcohol}
	out := [5]Axis{}
	for i, v := range axes {
		a, err := NewAxis(v)
		if err != nil {
			return Characteristics{}, err
		}
		out[i] = a
	}
	return Characteristics{
		Body:       out[0],
		Tannin:     out[1],
		Acidity:    out[2],
		Sweetness:  out[3],
		Alcohol:    out[4],
		Notes:      notes,
		Provenance: p,
	}, nil
}

// IsZero reports whether a Variety has no characteristics set yet (the axes are
// at their zero value, off the 1..5 rubric). Used by the seed-merge to treat an
// absent bundle the same as a default one.
func (c Characteristics) IsZero() bool { return c.Body == 0 }

// IsConfirmed reports whether the bundle has been vetted by the Drinker and so
// must survive a re-seed.
func (c Characteristics) IsConfirmed() bool { return c.Provenance == ProvenanceConfirmed }
