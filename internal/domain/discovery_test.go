package domain_test

import (
	"testing"

	"go-wine/internal/domain"
)

// chars is a test helper building a Characteristics bundle without going through
// the validating constructor's error return, so fixtures stay terse.
func chars(body, tannin, acidity, sweetness, alcohol int, notes ...string) domain.Characteristics {
	c, err := domain.NewCharacteristics(body, tannin, acidity, sweetness, alcohol, notes, domain.ProvenanceDefault)
	if err != nil {
		panic(err)
	}
	return c
}

// Discovery ranks untried Varieties by proximity to the enjoyed-grape SET: the
// candidate sitting in the same corner of characteristics space as an enjoyed
// grape ranks above one far from every enjoyed grape.
func TestDiscovery_RanksByProximityToSet(t *testing.T) {
	enjoyed := domain.ID("nebbiolo")
	profile := domain.TasteProfile{Enjoyed: []domain.EnjoyedVariety{{VarietyID: enjoyed, Weight: 5}}}
	enjoyedChars := map[domain.ID]domain.Characteristics{
		enjoyed: chars(5, 5, 4, 1, 4), // bold tannic red
	}

	near := domain.ID("aglianico")
	far := domain.ID("pinot-grigio")
	candidates := map[domain.ID]domain.Characteristics{
		near: chars(5, 5, 4, 1, 4), // identical to Nebbiolo
		far:  chars(1, 1, 4, 1, 1), // light, no tannin — a crisp white
	}

	recs := domain.Discover(profile, enjoyedChars, candidates)

	if len(recs) != 2 {
		t.Fatalf("want 2 recommendations, got %d: %+v", len(recs), recs)
	}
	if recs[0].VarietyID != near {
		t.Errorf("nearest candidate should rank first; got %q", recs[0].VarietyID)
	}
	if recs[1].VarietyID != far {
		t.Errorf("far candidate should rank last; got %q", recs[1].VarietyID)
	}
}

// A multimodal palate — loving both crisp light whites AND bold tannic reds —
// must get recommendations near EACH cluster and never in the medium-everything
// dead zone between them. This is the property a centroid would break: the
// average of the two clusters lands exactly on the dead-zone wine, which
// nearest-neighbour-to-the-set must rank WORST, not best.
func TestDiscovery_MultimodalAvoidsDeadZone(t *testing.T) {
	white := domain.ID("riesling")
	red := domain.ID("nebbiolo")
	profile := domain.TasteProfile{Enjoyed: []domain.EnjoyedVariety{
		{VarietyID: white, Weight: 5},
		{VarietyID: red, Weight: 5},
	}}
	enjoyedChars := map[domain.ID]domain.Characteristics{
		white: chars(1, 1, 5, 1, 1), // crisp light white
		red:   chars(5, 5, 2, 1, 5), // bold tannic red
	}

	nearWhite := domain.ID("gruner")
	nearRed := domain.ID("aglianico")
	deadZone := domain.ID("medium-everything")
	candidates := map[domain.ID]domain.Characteristics{
		nearWhite: chars(1, 1, 5, 1, 2), // hugs the white cluster
		nearRed:   chars(5, 4, 2, 1, 5), // hugs the red cluster
		deadZone:  chars(3, 3, 3, 1, 3), // the average — liked by neither
	}

	recs := domain.Discover(profile, enjoyedChars, candidates)

	if len(recs) != 3 {
		t.Fatalf("want 3 recommendations, got %d", len(recs))
	}
	// The dead-zone wine — nearest to the centroid — must rank LAST under
	// nearest-neighbour-to-the-set.
	if recs[2].VarietyID != deadZone {
		t.Errorf("dead-zone wine must rank worst, got order %q, %q, %q", recs[0].VarietyID, recs[1].VarietyID, recs[2].VarietyID)
	}
	// Both cluster-hugging candidates rank above the dead zone.
	top := map[domain.ID]bool{recs[0].VarietyID: true, recs[1].VarietyID: true}
	if !top[nearWhite] || !top[nearRed] {
		t.Errorf("both cluster candidates should outrank the dead zone; got %q, %q", recs[0].VarietyID, recs[1].VarietyID)
	}
}

// Each recommendation is justified by the specific enjoyed grape(s) it sits
// nearest to — explainability, not a bonus. A candidate hugging the red cluster
// is explained by the enjoyed red, not the enjoyed white.
func TestDiscovery_ExplainsByNearestEnjoyedGrape(t *testing.T) {
	white := domain.ID("riesling")
	red := domain.ID("nebbiolo")
	profile := domain.TasteProfile{Enjoyed: []domain.EnjoyedVariety{
		{VarietyID: white, Weight: 5},
		{VarietyID: red, Weight: 5},
	}}
	enjoyedChars := map[domain.ID]domain.Characteristics{
		white: chars(1, 1, 5, 1, 1),
		red:   chars(5, 5, 2, 1, 5),
	}
	nearRed := domain.ID("aglianico")
	candidates := map[domain.ID]domain.Characteristics{
		nearRed: chars(5, 4, 2, 1, 5),
	}

	recs := domain.Discover(profile, enjoyedChars, candidates)

	if len(recs) != 1 {
		t.Fatalf("want 1 recommendation, got %d", len(recs))
	}
	if len(recs[0].Because) != 1 || recs[0].Because[0] != red {
		t.Errorf("recommendation should be justified by the enjoyed red; got %+v", recs[0].Because)
	}
}

// An empty Taste profile yields no recommendations — the caller surfaces the
// explanatory empty state rather than an arbitrary list.
func TestDiscovery_EmptyProfileYieldsNothing(t *testing.T) {
	recs := domain.Discover(domain.TasteProfile{}, nil, map[domain.ID]domain.Characteristics{
		"x": chars(3, 3, 3, 1, 3),
	})
	if len(recs) != 0 {
		t.Errorf("empty profile should yield no recommendations; got %+v", recs)
	}
}

// Flavour-note tags combine as OVERLAP, not as a magnitude. Two candidates
// identical on the scalar axes are separated by how many flavour notes they
// share with the enjoyed grape: the one sharing more notes ranks nearer. This is
// the categorical operation kept distinct from the axis distance.
func TestDiscovery_TagOverlapBreaksAxisTies(t *testing.T) {
	enjoyed := domain.ID("nebbiolo")
	profile := domain.TasteProfile{Enjoyed: []domain.EnjoyedVariety{{VarietyID: enjoyed, Weight: 5}}}
	enjoyedChars := map[domain.ID]domain.Characteristics{
		enjoyed: chars(5, 5, 4, 1, 4, "cherry", "leather", "tar"),
	}

	sharesNotes := domain.ID("aglianico")
	noNotes := domain.ID("syrah")
	candidates := map[domain.ID]domain.Characteristics{
		// Identical axes; only the flavour-note overlap differs.
		sharesNotes: chars(5, 5, 4, 1, 4, "cherry", "leather"),
		noNotes:     chars(5, 5, 4, 1, 4, "banana", "bubblegum"),
	}

	recs := domain.Discover(profile, enjoyedChars, candidates)

	if recs[0].VarietyID != sharesNotes {
		t.Errorf("candidate sharing flavour notes should rank nearer; got %q first", recs[0].VarietyID)
	}
}
