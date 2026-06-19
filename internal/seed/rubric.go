// Package seed holds the conventional seed data shared by every store, so the
// Postgres and in-memory adapters seed identical reference values. Keeping it in
// one place is what lets Discovery's distances be meaningful: every Variety is
// scored on the same rubric with the same meaning.
package seed

import "go-wine/internal/app"

// Characteristics is the Variety-characteristics rubric: one coherent pass over
// the seeded grapes, every axis on the fixed 1..5 scale with one shared meaning
// (1 = lowest/lightest, 5 = highest/fullest). The grape names match the
// varieties seeded in migration 0002 (and seedMemory), which is the join key.
//
// The scale, anchored:
//   - body:      1 = water-light (Pinot Grigio)        … 5 = full/heavy (Syrah)
//   - tannin:    1 = none (whites)                      … 5 = grippy (Nebbiolo, Cabernet)
//   - acidity:   1 = soft/flat                          … 5 = razor-sharp (Riesling, Sauvignon Blanc)
//   - sweetness: 1 = bone-dry                           … 5 = sweet (none here are high; dry table wines)
//   - alcohol:   1 = low (~11%)                         … 5 = high (~15%, Grenache, Syrah)
//
// Whites carry tannin 1 by construction — tannin barely applies to white
// grapes, a standing property of the space, not a value to fudge.
func Characteristics() []app.VarietySeed {
	return []app.VarietySeed{
		// Reds.
		{Name: "Cabernet Sauvignon", Body: 5, Tannin: 5, Acidity: 3, Sweetness: 1, Alcohol: 4, Notes: []string{"blackcurrant", "cedar", "tobacco"}},
		{Name: "Merlot", Body: 4, Tannin: 3, Acidity: 3, Sweetness: 1, Alcohol: 4, Notes: []string{"plum", "chocolate"}},
		{Name: "Pinot Noir", Body: 2, Tannin: 2, Acidity: 4, Sweetness: 1, Alcohol: 3, Notes: []string{"cherry", "raspberry", "earth"}},
		{Name: "Syrah", Body: 5, Tannin: 4, Acidity: 3, Sweetness: 1, Alcohol: 5, Notes: []string{"blackberry", "pepper", "smoke"}},
		{Name: "Grenache", Body: 4, Tannin: 3, Acidity: 2, Sweetness: 1, Alcohol: 5, Notes: []string{"strawberry", "spice"}},
		{Name: "Mourvèdre", Body: 4, Tannin: 4, Acidity: 3, Sweetness: 1, Alcohol: 4, Notes: []string{"blackberry", "game", "earth"}},
		{Name: "Tempranillo", Body: 4, Tannin: 3, Acidity: 3, Sweetness: 1, Alcohol: 4, Notes: []string{"cherry", "leather", "tobacco"}},
		{Name: "Sangiovese", Body: 3, Tannin: 4, Acidity: 5, Sweetness: 1, Alcohol: 4, Notes: []string{"cherry", "herb", "leather"}},
		{Name: "Nebbiolo", Body: 4, Tannin: 5, Acidity: 5, Sweetness: 1, Alcohol: 4, Notes: []string{"rose", "tar", "cherry"}},
		{Name: "Malbec", Body: 4, Tannin: 4, Acidity: 3, Sweetness: 1, Alcohol: 4, Notes: []string{"blackberry", "plum", "violet"}},
		// Whites — tannin 1 by construction.
		{Name: "Chardonnay", Body: 4, Tannin: 1, Acidity: 3, Sweetness: 1, Alcohol: 4, Notes: []string{"apple", "butter", "vanilla"}},
		{Name: "Sauvignon Blanc", Body: 2, Tannin: 1, Acidity: 5, Sweetness: 1, Alcohol: 3, Notes: []string{"gooseberry", "grass", "citrus"}},
		{Name: "Riesling", Body: 2, Tannin: 1, Acidity: 5, Sweetness: 3, Alcohol: 2, Notes: []string{"lime", "petrol", "honey"}},
		{Name: "Pinot Grigio", Body: 1, Tannin: 1, Acidity: 4, Sweetness: 1, Alcohol: 3, Notes: []string{"pear", "lemon", "almond"}},
		{Name: "Chenin Blanc", Body: 3, Tannin: 1, Acidity: 4, Sweetness: 2, Alcohol: 3, Notes: []string{"quince", "honey", "apple"}},
	}
}
