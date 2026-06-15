# Wine Tracking

A personal web application for logging the wine I drink, understanding which
grapes I like (and under what circumstances), and discovering new ones.

## Language

**Variety**:
A grape variety — Shiraz, Pinot Noir, Riesling. The portable axis along which
taste is learned and discovery happens. I never log a Variety directly; my
preference for one is *derived* from the Wines I drink that contain it.
_Avoid_: Varietal (as a noun), grape type.

**Wine**:
A specific producer's product, identified by **producer + name + Style** and
*year-agnostic* — e.g. *Penfolds Bin 28 Shiraz*, not the 2019 specifically. Stable
reference data. Has a **Composition** of one or more Varieties. A single-grape
wine is the trivial case of one Variety at 100%. The vintage (year) is not part
of identity; it lives on the Tasting.

**Style**:
The canonical name for the label-level designation that implies a Wine's grapes,
whatever form the label takes — a grape ("Shiraz"), a place ("Chianti",
"Chablis"), or a blend name ("GSM", "Bordeaux blend"). A Style resolves to a
default **Composition** (seeded conventional knowledge, overridable). It is how a
developing drinker enters grapes they don't independently know.
_Avoid_: Appellation (too Old-World), type, category.

**Composition**:
The set of Varieties that make up a Wine, with rough proportions (even if only
"mostly Shiraz"). A GSM has Grenache + Shiraz + Mourvèdre; a Riesling has one
Variety at 100%.

**Tasting**:
A single logged event of drinking a Wine — the central thing I create. Its
explicit fields are the date, the wine's vintage (year), my **rating** — a
5-point measure of *absolute enjoyment* ("how much did I like it", never
value-for-money) — and any **Companions**. Everything else about the occasion
(food, mood, weather, anecdote) goes in a single free-text **Note**. The rating
lives here and *only* here: "do I like this Wine?" is always computed by
aggregating its Tastings, never stored on the Wine.
_Avoid_: Review, check-in.

**Note**:
The free-text capture of a Tasting's occasion — the canonical, lossless record
of context. Structured context facets (weather, food category, mood) are *not*
entered; they are derived from the Note by extraction, and only if/when
contextual discovery is built. The Note is the source of truth; any derived
structure is a regenerable cache.

**Companion**:
A named person a Drinker was with for a Tasting. Personal-zone reference data —
just a name — attachable to many of that Drinker's Tastings, so they can recall
which Wines a given person was around for and enjoyed. **Never a Drinker**: a
Companion is a name in the owner's personal zone, even if that same person also
happens to be a Drinker. The two never link, which is what keeps the app free of
any cross-Drinker sharing or consent.

**Drinker**:
The owner of a personal zone — Tastings, ratings, Companions, Taste profile all
belong to one. Several exist from the start; the app holds an **active Drinker**
chosen through a plain switcher (profile-selection, not sign-in). There is no
authentication yet, so a Drinker is an *identity for ownership*, not a secured
account.
_Avoid_: User (implies accounts/login, which we deliberately don't have),
account, profile (reserved for Taste profile).

**Variety characteristics**:
Intrinsic, conventional reference data about a grape — body, tannin, acidity,
sweetness, alcohol (scalar axes) and typical flavour notes (tags). The same
regardless of who drinks it, and it covers grapes I've never tried, which is
what lets Discovery reach beyond my own history. Seeded neutrally, **provenance**
-tagged, and only rarely hand-edited.
_Avoid_: Profile (that word is reserved for Taste profile), tasting notes.

**Provenance**:
The origin of a seeded value — both Variety characteristics and Composition
defaults carry it. Minimal and binary for now: *default* (neutral conventional
guess) vs *confirmed/edited by me*. Lets the app distinguish a guess from
grounded knowledge and never clobber a vetted value on re-seed. The richer
"which external source, on a trust ladder" story is a seam left for when a real
second source exists — not modelled yet.

**Taste profile**:
The *set* of grapes I've rated highly across my Tastings, each weighted by
enjoyment — my palate represented as a region of characteristics space (possibly
several clusters), **not** a single average point. Derived from enjoyment,
sharpens as I log more, and is the anchor Discovery recommends *near*. Distinct
from Variety characteristics (intrinsic to a grape) — this one is intrinsic to
*me*.

**Discovery**:
Recommending Varieties I haven't logged that sit near my Taste profile in
characteristics space. The app's third purpose: finding genuinely new grapes,
not just re-surfacing ones I know.

## Flagged ambiguities

- "Variety" vs "Wine": a GSM is one **Wine** with three **Varieties** in its
  **Composition**, not three things I drank.
