package domain

import (
	"testing"
	"time"
)

// mustRating builds a Rating in tests, failing loudly on a bad value.
func mustRating(t *testing.T, v int) Rating {
	t.Helper()
	r, err := NewRating(v)
	if err != nil {
		t.Fatalf("NewRating(%d): %v", v, err)
	}
	return r
}

// tastingOf is a minimal Tasting fixture for a Drinker's rating of a Wine.
func tastingOf(t *testing.T, drinkerID, wineID ID, rating int) Tasting {
	t.Helper()
	return Tasting{
		ID:        NewID(),
		DrinkerID: drinkerID,
		WineID:    wineID,
		Rating:    mustRating(t, rating),
		DrunkOn:   time.Now(),
	}
}

// LikeWine aggregates a single Wine's Tastings into a verdict — the mean of the
// ratings and the count. No tastings yields a not-yet-tasted verdict, never a
// fake zero.
func TestLikeWine_MeanOfRatings(t *testing.T) {
	drinker := NewID()
	wine := NewID()
	tastings := []Tasting{
		tastingOf(t, drinker, wine, 4),
		tastingOf(t, drinker, wine, 5),
	}

	got := LikeWine(tastings)

	if !got.Tasted {
		t.Fatalf("two tastings should count as tasted; got %+v", got)
	}
	if got.Count != 2 {
		t.Errorf("Count = %d, want 2", got.Count)
	}
	if got.MeanRating != 4.5 {
		t.Errorf("MeanRating = %v, want 4.5", got.MeanRating)
	}
}

// A Wine with no Tastings is not-yet-tasted — a clear unknown, not a zero.
func TestLikeWine_NoTastingsIsNotTasted(t *testing.T) {
	got := LikeWine(nil)
	if got.Tasted {
		t.Errorf("no tastings should be not-yet-tasted; got %+v", got)
	}
	if got.Count != 0 || got.MeanRating != 0 {
		t.Errorf("not-yet-tasted should carry no rating; got %+v", got)
	}
}

// part is a CompositionPart fixture.
func part(varietyID ID, proportion int) CompositionPart {
	return CompositionPart{VarietyID: varietyID, Proportion: proportion}
}

// LikeVariety attributes a Drinker's enjoyment to a Variety through the
// Compositions of the wines drunk: each Tasting's rating flows to the wine's
// Varieties weighted by their proportion of the blend. For a single-grape wine
// the Variety carries the full rating.
func TestLikeVariety_SingleGrapeWineCarriesFullRating(t *testing.T) {
	drinker := NewID()
	shiraz := NewID()
	wine := NewID()
	comps := map[ID]Composition{
		wine: {Parts: []CompositionPart{part(shiraz, 100)}},
	}
	tastings := []Tasting{
		tastingOf(t, drinker, wine, 4),
		tastingOf(t, drinker, wine, 5),
	}

	got := LikeVariety(shiraz, tastings, comps)

	if !got.Tasted {
		t.Fatalf("a drunk single-grape wine should make the Variety tasted; got %+v", got)
	}
	if got.Preference != 4.5 {
		t.Errorf("Preference = %v, want 4.5 (full rating attributed)", got.Preference)
	}
}

// In a blend the rating is attributed to each Variety weighted by its
// proportion: the preference is the proportion-weighted mean rating across all
// wines drunk that contain the Variety.
func TestLikeVariety_BlendAttributesByProportion(t *testing.T) {
	drinker := NewID()
	grenache := NewID()
	syrah := NewID()
	mourvedre := NewID()
	gsm := NewID()       // 50 Grenache / 30 Syrah / 20 Mourvèdre, rated 5
	soloSyrah := NewID() // 100 Syrah, rated 1
	comps := map[ID]Composition{
		gsm:       {Parts: []CompositionPart{part(grenache, 50), part(syrah, 30), part(mourvedre, 20)}},
		soloSyrah: {Parts: []CompositionPart{part(syrah, 100)}},
	}
	tastings := []Tasting{
		tastingOf(t, drinker, gsm, 5),
		tastingOf(t, drinker, soloSyrah, 1),
	}

	// Grenache appears only in the GSM at 50%: preference = 5 (a weighted mean of
	// the one contribution).
	gren := LikeVariety(grenache, tastings, comps)
	if gren.Preference != 5 {
		t.Errorf("Grenache Preference = %v, want 5", gren.Preference)
	}

	// Syrah appears in both: GSM at 0.30 weight (rating 5) and the solo at 1.0
	// weight (rating 1). Proportion-weighted mean = (5*0.30 + 1*1.0)/(0.30+1.0)
	// = 2.5/1.3 ≈ 1.923.
	syr := LikeVariety(syrah, tastings, comps)
	want := (5*0.30 + 1*1.0) / (0.30 + 1.0)
	if abs(syr.Preference-want) > 1e-9 {
		t.Errorf("Syrah Preference = %v, want %v", syr.Preference, want)
	}
}

// A Variety the Drinker has never drunk (no wine containing it tasted) is
// not-yet-tasted — a clear unknown, not a zero preference.
func TestLikeVariety_NeverDrunkIsNotTasted(t *testing.T) {
	drinker := NewID()
	shiraz := NewID()
	riesling := NewID()
	wine := NewID()
	comps := map[ID]Composition{
		wine: {Parts: []CompositionPart{part(shiraz, 100)}},
	}
	tastings := []Tasting{tastingOf(t, drinker, wine, 5)}

	got := LikeVariety(riesling, tastings, comps)
	if got.Tasted {
		t.Errorf("a never-drunk Variety should be not-yet-tasted; got %+v", got)
	}
}

func abs(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}

// hasVariety reports whether the profile contains an enjoyed entry for the
// given Variety, returning its weight.
func enjoyedWeight(p TasteProfile, id ID) (float64, bool) {
	for _, e := range p.Enjoyed {
		if e.VarietyID == id {
			return e.Weight, true
		}
	}
	return 0, false
}

// The Taste profile is the SET of highly-rated Varieties (derived through the
// same Composition attribution), each weighted by enjoyment. A Variety rated at
// or above the enjoyed threshold is included; one below is left out.
func TestTasteProfile_IncludesEnjoyedVarietiesAboveThreshold(t *testing.T) {
	drinker := NewID()
	shiraz := NewID()
	pinot := NewID()
	bigRed := NewID() // 100 Shiraz, loved (5)
	mehRed := NewID() // 100 Pinot, disliked (2)
	comps := map[ID]Composition{
		bigRed: {Parts: []CompositionPart{part(shiraz, 100)}},
		mehRed: {Parts: []CompositionPart{part(pinot, 100)}},
	}
	tastings := []Tasting{
		tastingOf(t, drinker, bigRed, 5),
		tastingOf(t, drinker, mehRed, 2),
	}

	profile := TasteProfile{}.Build(tastings, comps)

	w, ok := enjoyedWeight(profile, shiraz)
	if !ok {
		t.Fatalf("a loved grape should be in the profile; got %+v", profile)
	}
	if w != 5 {
		t.Errorf("Shiraz weight = %v, want 5 (its enjoyment)", w)
	}
	if _, ok := enjoyedWeight(profile, pinot); ok {
		t.Errorf("a disliked grape should not be in the profile; got %+v", profile)
	}
}

// The crux of the issue: a multimodal palate (loving two well-separated regions
// of characteristics space — crisp whites AND bold tannic reds) stays TWO
// regions. The profile is kept as a set of enjoyed grapes, never collapsed to a
// single average point, so both clusters survive.
func TestTasteProfile_MultimodalPalateKeepsBothClusters(t *testing.T) {
	drinker := NewID()
	// Cluster A — bold tannic reds.
	shiraz := NewID()
	cabernet := NewID()
	// Cluster B — crisp light whites.
	riesling := NewID()
	sauvignon := NewID()

	bigRed1 := NewID()
	bigRed2 := NewID()
	crispWhite1 := NewID()
	crispWhite2 := NewID()
	comps := map[ID]Composition{
		bigRed1:     {Parts: []CompositionPart{part(shiraz, 100)}},
		bigRed2:     {Parts: []CompositionPart{part(cabernet, 100)}},
		crispWhite1: {Parts: []CompositionPart{part(riesling, 100)}},
		crispWhite2: {Parts: []CompositionPart{part(sauvignon, 100)}},
	}
	tastings := []Tasting{
		tastingOf(t, drinker, bigRed1, 5),
		tastingOf(t, drinker, bigRed2, 5),
		tastingOf(t, drinker, crispWhite1, 5),
		tastingOf(t, drinker, crispWhite2, 4),
	}

	profile := TasteProfile{}.Build(tastings, comps)

	// All four enjoyed grapes are present, as a set — both clusters preserved.
	for _, id := range []ID{shiraz, cabernet, riesling, sauvignon} {
		if _, ok := enjoyedWeight(profile, id); !ok {
			t.Errorf("multimodal profile should keep grape %s; got %+v", id, profile)
		}
	}
	if len(profile.Enjoyed) != 4 {
		t.Errorf("profile should hold all four enjoyed grapes (a set/region, not an average); got %d", len(profile.Enjoyed))
	}
}

// An empty history yields an empty profile, not a fabricated one.
func TestTasteProfile_EmptyHistoryIsEmpty(t *testing.T) {
	profile := TasteProfile{}.Build(nil, nil)
	if len(profile.Enjoyed) != 0 {
		t.Errorf("no tastings should yield an empty profile; got %+v", profile)
	}
}
