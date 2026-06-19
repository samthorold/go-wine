package domain

// MergeComposition is the provenance seed-merge for a Wine's Composition — the
// sibling of MergeCharacteristics, applying the same inviolable rule to the
// Style → default Composition seed:
//
//	a re-seed never clobbers a confirmed value.
//
// So:
//   - if the stored Composition is confirmed-by-me (the Drinker named or edited
//     the grapes), it is preserved untouched and the Style seed is ignored;
//   - otherwise (the stored Composition is absent or still a Style-default guess)
//     the seeded default applies.
//
// Re-running the Style seed is therefore both idempotent and non-clobbering:
// confirmed wine compositions survive every re-seed, while default ones track the
// current Style seed.
func MergeComposition(seeded, stored Composition) Composition {
	if stored.IsConfirmed() {
		return stored
	}
	return seeded
}
