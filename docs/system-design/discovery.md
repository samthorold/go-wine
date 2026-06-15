# Discovery

How the app finds new grape Varieties for a drinker to try. This is the app's
reason to exist beyond a logbook, and it shapes the rest of the model.

## The mechanism

Every Variety has a position in **characteristics space** — the scalar axes
(body, tannin, acidity, sweetness, alcohol) plus flavour-note tags. A drinker's
**Taste profile** is the *set* of grapes they have rated highly, each weighted by
enjoyment. Discovery ranks the Varieties the drinker has *not* logged by their
proximity to that **set** — nearest-neighbour against the enjoyed grapes — and
surfaces the nearest, each annotated with the enjoyed grape(s) that justify it.

## Why proximity-to-a-set, not to an average

The profile is a set, not a centre-of-gravity, because a developing palate is
usually **multimodal**: a drinker comes to love crisp light whites *and* big
tannic reds, which sit at opposite corners of characteristics space. Their
average lands in the empty middle — a medium-everything wine representing neither
cluster, and quite possibly disliked. A centroid would confidently recommend
exactly those dead-zone wines. Nearest-neighbour against the set recommends near
*each* cluster and never between them.

It also makes every recommendation **explainable**: it arrives attached to the
specific enjoyed grape(s) it sits near ("Aglianico, because you loved Nebbiolo
and Sangiovese"). A centroid can state no such reason. Explainability is not a
bonus here — it is the same palate-vocabulary teaching the app exists to do.

The enjoyment weight is the Tasting's 5-point rating, and it must measure
*absolute* enjoyment, not value-for-money. A relative rating would pull the
profile toward bargains rather than taste — two equally-enjoyed wines would carry
different weights purely because of price — so any value judgement is kept out of
the rating entirely.

Scalar axes combine as distance; flavour-note tags combine as overlap. They are
not the same operation because tags are categorical — "tastes of cherry and
leather" is set membership, not a magnitude — so collapsing them into a single
distance would be a category error.

## Why characteristics are intrinsic reference data, not derived

The whole point is to recommend grapes the drinker has **never drunk**. Their
characteristics therefore cannot come from the drinker's own Tastings — there
are none for an undrunk grape. The knowledge must exist independently of any
Tasting, covering grapes the drinker hasn't met. So Variety characteristics are
reference data seeded from conventional wine knowledge, kept strictly separate
from Tastings (which carry preference, not chemistry).

A tempting alternative — derive each grape's characteristics from the Tastings
of wines that contain it — fails precisely on the grapes that matter: it can
only ever describe grapes already drunk, the exact ones that don't need
recommending.

## Why the seed is a neutral prior, corrected by enjoyment — not by editing

The seed is conventional and neutral, because the drinker we design for is still
learning and has no expert palate to anchor it to. That same fact rules out the
obvious correction loop — "let the drinker fix wrong characteristics over time."
Editing characteristics demands exactly the palate vocabulary a developing
drinker lacks; a correction loop that depends on it never fires, and the seed
freezes forever.

So correction runs on the one signal a developing drinker reliably produces:
**enjoyment.** They always know whether they liked a wine; they don't reliably
know it was the tannin. Every Tasting — including recommended wines they go try
— pulls the Taste profile toward what they actually enjoy. Where the drinker has
real data, the seed's errors wash out; where they don't, the seed's structure
still lets Discovery reach new grapes. Hand-editing characteristics remains as a
rare escape hatch, **Provenance**-tagged so a vetted value is never overwritten
on re-seed, but it is not the mechanism the design relies on.

A consequence the design leans into: because the profile is expressed in named
axes, the app can *report the profile back* ("you consistently rate high-acid,
light-bodied reds well"), teaching the drinker the vocabulary they're missing
rather than demanding it up front.

## The input side needs its own seed

The chain is only as strong as its weakest link, and for a developing drinker
that link is **Wine → Composition**. They frequently won't know a wine's grapes,
and Old World bottles are labelled by place, not grape ("Chianti," "Côtes du
Rhône"), where the composition is implied by convention but never stated. An
empty Composition severs the chain — a Tasting that connects to no Variety
teaches the profile nothing.

So a *second* body of conventional knowledge is seeded: a **style/appellation →
typical Composition** map. Logging a wine, the drinker either names the grapes
(if known) or picks the label's style, and a default Composition fills in,
overridable. It carries the same minimal **Provenance** as characteristics — a
default is distinguishable from a confirmed value, and confirmations survive a
re-seed.

Provenance is deliberately binary (default vs confirmed-by-me), because there is
exactly one external source today: the conventional seed. A multi-source trust
ladder would be a one-rung hierarchy. The provenance field is the standing seam
where richer sources slot in later; ranking them is deferred until a second real
source exists.

## The seed must share one yardstick

Proximity is only meaningful if every Variety is scored on the same scale with
the same meaning — if "high tannin" drifts between grapes, distances are noise
and recommendations are random. The seed is therefore generated in one coherent
pass against a single explicit rubric (the axes, a fixed scale, and anchored
definitions of each scale point), not grape-by-grape in isolation.

## The searchable space is small, so v1 needs no index

Discovery searches over **Varieties** — a few hundred grapes. Ranking untried
grapes by distance to the enjoyed set is a computation over a few hundred rows,
done on each request with no vector index, embedding store, or precomputed
structure. The "indexing cost" that dominates recommendation systems does not
arise at this scale, and v1 deliberately spends nothing on it.

That cost appears only when the searchable space stops being grapes. Two
extensions, both deferred and both *additive* rather than re-architecting:

- **Contextual search** — Notes extracted into mood/food/companion facets that
  become searchable dimensions, so a query can be "something for a cosy night
  with Sarah." This is where real re-indexing begins: maintaining a derived,
  searchable representation of unstructured history. The capture design already
  accommodates it — Notes are canonical and facets are a regenerable cache, so
  "re-index" means "regenerate that cache into a search space," not rebuild.
- **Conversational exploration** — discovery as a dialogue ("something lighter
  than the Barolo") mediated against history, rather than a ranked list.

Neither is part of v1; both are reachable from this design without unwinding it.

## A standing limitation

Tannin barely applies to white grapes, so reds and whites occupy different
regions of characteristics space. Cross-colour recommendations are therefore
rare by construction — a property of the space, not a bug to patch.
