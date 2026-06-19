package domain

import (
	"math"
	"sort"
)

// Discovery is the app's reason to exist: recommend grape Varieties a Drinker
// has not logged, ranked by proximity to their Taste profile — the SET of
// enjoyed grapes — in characteristics space. It is a read-only domain service,
// pure and DB-free, taking already-fetched data and computing a ranking. It is
// unit-tested in isolation against hand-built fixtures.
//
// Proximity is to the SET, never to a centroid. For each untried candidate its
// distance to the profile is the distance to its NEAREST enjoyed grape (the
// minimum over the enjoyed set), not the distance to an averaged point. This is
// what makes a multimodal palate work: a Drinker who loves crisp whites AND bold
// tannic reds is recommended near BOTH clusters and never the medium-everything
// dead zone between them (which is exactly where a centroid would point).
//
// The two kinds of characteristic combine by DIFFERENT operations, never
// collapsed into one magnitude:
//
//   - Scalar axes (body, tannin, acidity, sweetness, alcohol) combine as a
//     numeric DISTANCE — Euclidean over the 1..5 rubric.
//   - Flavour-note tags combine as OVERLAP — Jaccard similarity of the two tag
//     sets (shared / union). Treating tags as a magnitude would be a category
//     error: "tastes of cherry and leather" is set membership, not a quantity.
//
// They are combined into a single comparable score per (candidate, enjoyed)
// pair: axisDistance + tagWeight·(1 − jaccard). The tag term is a penalty in
// [0, tagWeight] — 0 when the tag sets are identical, tagWeight when disjoint —
// so a shared flavour profile pulls a candidate closer without ever dominating
// the axis geometry. The candidate's score against the SET is the MINIMUM of
// that combined score over the enjoyed grapes (nearest-neighbour). Lower is
// nearer; recommendations are returned ascending by score.

// tagWeight scales the tag-overlap penalty into the same order of magnitude as
// the axis distance. The axis distance spans 0..√80 ≈ 8.94; tagWeight = 2 lets a
// fully-shared vs fully-disjoint flavour profile move a candidate by up to 2
// units — meaningful, but unable to overturn a large geometric gap (the axis
// difference between a crisp white and a bold red). Documented and deliberate.
const tagWeight = 2.0

// Recommendation is one ranked suggestion: an untried Variety, the combined
// distance (Score, lower = nearer) to the nearest enjoyed grape(s), and the
// enjoyed grape(s) that justify it — the explainability the design requires.
type Recommendation struct {
	VarietyID ID
	Score     float64
	// Because is the enjoyed grape(s) the candidate sits nearest to: the ones at
	// (within an epsilon of) the minimum distance. Usually one; several when the
	// candidate is equidistant from more than one enjoyed grape.
	Because []ID
}

// Discover ranks untried candidate Varieties by proximity to the enjoyed-grape
// set. profile is the Taste profile (the enjoyed grapes and their enjoyment
// weights); enjoyedChars gives each enjoyed grape's Characteristics; candidates
// gives each untried Variety's Characteristics. The result is ordered nearest
// first; each Recommendation carries the enjoyed grape(s) that justify it.
//
// A candidate with no characteristics (IsZero) cannot be placed in the space and
// is skipped. An enjoyed grape with no characteristics likewise contributes no
// position. An empty profile yields no recommendations — the caller surfaces the
// explanatory empty state.
func Discover(profile TasteProfile, enjoyedChars map[ID]Characteristics, candidates map[ID]Characteristics) []Recommendation {
	var recs []Recommendation
	for candID, candChars := range candidates {
		if candChars.IsZero() {
			continue
		}
		best := math.Inf(1)
		var because []ID
		for _, e := range profile.Enjoyed {
			ec, ok := enjoyedChars[e.VarietyID]
			if !ok || ec.IsZero() {
				continue
			}
			d := pairDistance(candChars, ec)
			switch {
			case d < best-distanceEpsilon:
				best = d
				because = []ID{e.VarietyID}
			case d <= best+distanceEpsilon:
				because = append(because, e.VarietyID)
			}
		}
		if math.IsInf(best, 1) {
			continue // no enjoyed grape could be placed against this candidate
		}
		recs = append(recs, Recommendation{VarietyID: candID, Score: best, Because: because})
	}
	sort.Slice(recs, func(i, j int) bool {
		if math.Abs(recs[i].Score-recs[j].Score) > distanceEpsilon {
			return recs[i].Score < recs[j].Score
		}
		return recs[i].VarietyID < recs[j].VarietyID
	})
	return recs
}

// distanceEpsilon is the tolerance within which two distances count as equal,
// both for ranking tie-breaks and for collecting co-nearest justifying grapes.
const distanceEpsilon = 1e-9

// pairDistance is the combined distance between two grapes: Euclidean over the
// scalar axes plus the tag-overlap penalty. The two operations stay distinct —
// distance for axes, overlap for tags — combined only at the end.
func pairDistance(a, b Characteristics) float64 {
	return axisDistance(a, b) + tagWeight*(1-jaccard(a.Notes, b.Notes))
}

// axisDistance is the Euclidean distance between two grapes over the five 1..5
// scalar axes.
func axisDistance(a, b Characteristics) float64 {
	sq := func(x int) float64 { return float64(x) * float64(x) }
	sum := sq(a.Body.Int()-b.Body.Int()) +
		sq(a.Tannin.Int()-b.Tannin.Int()) +
		sq(a.Acidity.Int()-b.Acidity.Int()) +
		sq(a.Sweetness.Int()-b.Sweetness.Int()) +
		sq(a.Alcohol.Int()-b.Alcohol.Int())
	return math.Sqrt(sum)
}

// jaccard is the overlap of two flavour-note tag sets: |intersection| / |union|.
// Two empty sets share nothing to compare, so they score 0 (no overlap evidence)
// rather than a spurious 1.
func jaccard(a, b []string) float64 {
	if len(a) == 0 && len(b) == 0 {
		return 0
	}
	set := make(map[string]bool, len(a))
	for _, n := range a {
		set[n] = true
	}
	union := make(map[string]bool, len(a)+len(b))
	for n := range set {
		union[n] = true
	}
	inter := 0
	for _, n := range b {
		if set[n] {
			inter++
		}
		union[n] = true
	}
	return float64(inter) / float64(len(union))
}
