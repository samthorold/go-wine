# Capturing Tastings

How an occasion of drinking is recorded. The guiding constraint is friction: a
Tasting is logged on a phone, after a glass, by someone who'll stop logging
entirely if it takes effort. So capture asks for the least that preserves the
signal.

## Free text is the canonical context; structure is derived

A Tasting's context — food, mood, weather, anecdote — is captured as a single
free-text **Note**, not a form of pickers. The Note is the source of truth.
Structured facets (weather bucket, food category, mood tags) are *derived* from
it by extraction, materialised only if and when contextual discovery is built,
in whatever schema is decided at that point.

This inverts the usual "capture structure at input time or lose it forever"
rule, because that rule assumes structure can only be recorded as it happens.
It can't be un-captured here: a Note is richer than any fixed form, and
extraction can turn the whole history into any schema retroactively. A form, by
contrast, throws away everything its buckets didn't anticipate and forces the
schema to be guessed now, before the contextual-discovery feature that would
define it even exists.

Two standing consequences:

- Derived facets are a **cache, never truth** — extraction is non-deterministic
  and occasionally wrong, so structured context is always regenerable from the
  Note and never authoritative over it.
- A Note only contains what the drinker chose to write, so context is captured
  more spottily than a prompting form would yield. This is the accepted cost of
  low friction.

## What stays explicit, and why

Two things are *not* folded into the Note:

- **Rating** — the 5-point enjoyment score. It is the deterministic weight the
  entire Taste profile math depends on; inferring it from prose would make the
  load-bearing signal non-deterministic for no benefit, when it costs one tap.
- **Companions** — named reference entities aggregated across Tastings ("which
  Wines was Sarah around for?"). Resolving a person from free text is
  error-prone entity-matching; an explicit pick keeps the identity clean.

Everything else about the occasion is Note.
