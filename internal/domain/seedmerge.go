package domain

// MergeCharacteristics is the provenance seed-merge — one of the few genuine
// command-side domain rules in the app. Given the value a coherent seed pass
// would set (seeded) and whatever is currently stored for a Variety (stored), it
// yields the value to persist, with one inviolable rule:
//
//	a re-seed never clobbers a confirmed value.
//
// So:
//   - if the stored bundle is confirmed-by-me, it is preserved untouched and the
//     seed is ignored — this is the no-clobber rule that protects a vetted value;
//   - otherwise (the stored value is absent or still a default guess) the seeded
//     default applies.
//
// Re-running the seed is therefore both idempotent and non-clobbering: confirmed
// values survive every re-seed, while unconfirmed grapes track the current seed.
func MergeCharacteristics(seeded, stored Characteristics) Characteristics {
	if stored.IsConfirmed() {
		return stored
	}
	return seeded
}
