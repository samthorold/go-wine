package domain

// Preference read models — the query-side domain services that answer "do I
// like this Wine?", "do I like this Variety?", and build the enjoyment-weighted
// Taste profile. They are read-only: they take already-fetched data (Tastings,
// the Compositions of the wines drunk) and compute a result, never mutating an
// aggregate. This is the richest logic in the app and is unit-tested in
// isolation against in-memory fixtures with no DB.

// WineVerdict is the aggregate of one Wine's Tastings — "do I like this Wine?".
// The Rating lives only on the Tasting, so a Wine's verdict is always computed
// by aggregating its Tastings, never stored. Tasted is false when the Drinker
// has no Tastings of the Wine yet: a clear not-yet-tasted answer, not a fake
// zero rating.
type WineVerdict struct {
	Tasted     bool
	Count      int
	MeanRating float64
}

// LikeWine aggregates the given Tastings (the active Drinker's Tastings of one
// Wine) into a verdict: the mean of their ratings and the count. An empty slice
// yields the zero WineVerdict (Tasted=false).
func LikeWine(tastings []Tasting) WineVerdict {
	if len(tastings) == 0 {
		return WineVerdict{}
	}
	sum := 0
	for _, t := range tastings {
		sum += t.Rating.Int()
	}
	return WineVerdict{
		Tasted:     true,
		Count:      len(tastings),
		MeanRating: float64(sum) / float64(len(tastings)),
	}
}

// VarietyPreference is a Drinker's derived preference for a grape — "do I like
// this Variety?". A Variety is never logged directly; the preference is
// attributed through the Compositions of the wines drunk that contain it.
// Tasted is false when no wine containing the Variety has been tasted: a clear
// not-yet-tasted answer rather than a fake zero.
type VarietyPreference struct {
	Tasted     bool
	Preference float64
}

// LikeVariety attributes the Drinker's enjoyment to a Variety through the
// Compositions of the wines drunk. Attribution formula: each Tasting's rating
// contributes to the Variety weighted by the Variety's proportion of that
// wine's blend (proportion/100). The preference is the proportion-weighted mean
// rating across every wine drunk that contains the Variety — a single-grape
// wine carries the full rating, a minor blending grape carries little. A
// Variety in no drunk wine is not-yet-tasted.
//
// comps maps each WineID to its Composition; a Tasting of a wine missing from
// comps (no grapes known) contributes nothing — the broken Wine→Composition
// link teaches the profile nothing, exactly as the discovery design states.
func LikeVariety(varietyID ID, tastings []Tasting, comps map[ID]Composition) VarietyPreference {
	weightedSum := 0.0
	totalWeight := 0.0
	for _, t := range tastings {
		c, ok := comps[t.WineID]
		if !ok {
			continue
		}
		for _, p := range c.Parts {
			if p.VarietyID != varietyID {
				continue
			}
			w := float64(p.Proportion) / 100
			weightedSum += float64(t.Rating.Int()) * w
			totalWeight += w
		}
	}
	if totalWeight == 0 {
		return VarietyPreference{}
	}
	return VarietyPreference{Tasted: true, Preference: weightedSum / totalWeight}
}

// enjoyedThreshold is the derived preference (on the 1..5 absolute-enjoyment
// scale) at or above which a Variety counts as "enjoyed" and enters the Taste
// profile. 4.0 — "really liked, consistently" — keeps the profile to grapes the
// Drinker genuinely enjoys rather than merely tolerated.
const enjoyedThreshold = 4.0

// EnjoyedVariety is one grape in the Taste profile: a Variety the Drinker rates
// highly, with the enjoyment Weight (its derived preference) that Discovery's
// nearest-neighbour search uses to rank candidates.
type EnjoyedVariety struct {
	VarietyID ID
	Weight    float64
}

// TasteProfile is the Drinker's palate as the SET of grapes they rate highly,
// each weighted by enjoyment — a region of characteristics space, possibly
// several separated clusters, deliberately NOT collapsed to a single average
// point. A multimodal palate (crisp whites and bold reds at once) would average
// to a medium-everything dead zone liked by neither; keeping the set preserves
// every cluster, which is the property Discovery's proximity-to-a-set search
// depends on. This is the clean, reusable value #9 (Discovery) consumes.
type TasteProfile struct {
	Enjoyed []EnjoyedVariety
}

// Build computes the Taste profile from the Drinker's Tastings and the
// Compositions of the wines drunk. Every Variety reachable through those
// Compositions is scored by LikeVariety's proportion-weighted attribution; the
// ones at or above the enjoyed threshold enter the profile, each carrying its
// derived preference as the enjoyment weight. The result is a set/region — the
// enjoyed grapes are kept individually, never averaged together.
func (TasteProfile) Build(tastings []Tasting, comps map[ID]Composition) TasteProfile {
	// Collect the distinct Varieties reachable through the drunk wines, in a
	// deterministic order (first appearance) so the profile is stable.
	seen := make(map[ID]bool)
	var order []ID
	for _, t := range tastings {
		c, ok := comps[t.WineID]
		if !ok {
			continue
		}
		for _, p := range c.Parts {
			if !seen[p.VarietyID] {
				seen[p.VarietyID] = true
				order = append(order, p.VarietyID)
			}
		}
	}

	var enjoyed []EnjoyedVariety
	for _, id := range order {
		pref := LikeVariety(id, tastings, comps)
		if pref.Tasted && pref.Preference >= enjoyedThreshold {
			enjoyed = append(enjoyed, EnjoyedVariety{VarietyID: id, Weight: pref.Preference})
		}
	}
	return TasteProfile{Enjoyed: enjoyed}
}
