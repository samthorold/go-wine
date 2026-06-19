package seed

import "go-wine/internal/app"

// StyleCompositions is the input-side seed: a conventional Style/appellation →
// default Composition map, the sibling of the characteristics rubric on the
// output side. For a developing drinker the weakest link in the
// Wine → Composition → Variety chain is the grapes, because Old World bottles are
// labelled by place, not grape ("Chianti", "Chablis", "Côtes du Rhône"), where
// the composition is implied by convention but never stated. Picking the label's
// Style fills an overridable default Composition so a Tasting still reaches a
// Variety (see docs/system-design/discovery.md, "The input side needs its own
// seed").
//
// Each default is a conventional blend over the grapes seeded by name in
// migration 0002 / seedMemory — the name is the join key, since each store mints
// its own random Variety IDs. The proportions are deliberately rough ("mostly
// Sangiovese"), within the domain's sum-to-~100% tolerance, and Provenance is
// always 'default': the seed only ever proposes, and MergeComposition decides
// whether it lands (a confirmed Composition survives a re-seed).
//
// The rationale for each entry (a small, sane handful spanning Old World
// appellations and a New World varietal label):
//   - Chianti        — Tuscan red, Sangiovese-dominant by DOCG rule.
//   - GSM             — the southern-Rhône / Châteauneuf blend: Grenache-led,
//     with Syrah and Mourvèdre (the "GSM" initialism).
//   - Chablis         — northern-Burgundy white, 100% Chardonnay by appellation.
//   - Bordeaux Blend  — the classic left-bank claret: Cabernet-Sauvignon-led with
//     Merlot.
//   - Rioja           — Spain's flagship red, Tempranillo-dominant.
//   - Côtes du Rhône  — the broader Rhône red, Grenache-led like a softer GSM.
//
// Styles whose grapes a store happens not to carry resolve over whatever grapes
// it does have (the resolver joins by name and drops the rest); the entries
// above are chosen so the seeded grape set covers them.
func StyleCompositions() []app.StyleSeed {
	return []app.StyleSeed{
		{Style: "Chianti", Parts: []app.StylePartSeed{
			{Variety: "Sangiovese", Proportion: 85},
			{Variety: "Merlot", Proportion: 15},
		}},
		{Style: "GSM", Parts: []app.StylePartSeed{
			{Variety: "Grenache", Proportion: 50},
			{Variety: "Syrah", Proportion: 30},
			{Variety: "Mourvèdre", Proportion: 20},
		}},
		{Style: "Chablis", Parts: []app.StylePartSeed{
			{Variety: "Chardonnay", Proportion: 100},
		}},
		{Style: "Bordeaux Blend", Parts: []app.StylePartSeed{
			{Variety: "Cabernet Sauvignon", Proportion: 70},
			{Variety: "Merlot", Proportion: 30},
		}},
		{Style: "Rioja", Parts: []app.StylePartSeed{
			{Variety: "Tempranillo", Proportion: 90},
			{Variety: "Grenache", Proportion: 10},
		}},
		{Style: "Côtes du Rhône", Parts: []app.StylePartSeed{
			{Variety: "Grenache", Proportion: 60},
			{Variety: "Syrah", Proportion: 40},
		}},
	}
}
