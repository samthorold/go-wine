# System Design

How this app models wine tracking and discovery. These docs are
implementation-agnostic; the code encodes them. Language is defined once in
[`/CONTEXT.md`](../../CONTEXT.md) and used here without redefinition.

## The docs

- [discovery.md](./discovery.md) — the heart of the app: how new grape Varieties
  are recommended (nearest-neighbour over the enjoyed-grape set, explainable),
  why characteristics are intrinsic seeded reference data, and why correction is
  enjoyment-driven rather than by hand-editing.
- [capturing-tastings.md](./capturing-tastings.md) — why a Tasting's context is a
  single canonical free-text Note with structure derived later, and what stays
  explicit (rating, Companions).
- [data-ownership.md](./data-ownership.md) — the personal/reference zone
  partition, and multi-tenancy via a Drinker switcher without authentication.
- [hypermedia-ui.md](./hypermedia-ui.md) — why the UI is hypermedia-driven
  (server owns all markup and state, no client data API), required rather than
  JS-optional, a boosted multi-page app carved along swap boundaries, with
  failed commands re-rendering their form.
- [look-and-feel.md](./look-and-feel.md) — why styling is minimal, classless
  (Pico.css) and dark by default with no toggle, custom CSS rationed to domain
  accents, and no client-side UI components.

## The v1 line

**In v1:**
- Log **Tastings**: pick/create a **Wine** (producer + name + **Style**; Style
  seeds an overridable **Composition**), record vintage-on-the-tasting, a
  5-point enjoyment **rating**, any **Companions**, and a free-text **Note**.
- Two **seeded** reference bodies — Variety **characteristics** and
  Style→**Composition** — editable, each with binary **Provenance** (default vs
  confirmed; confirmations survive re-seed).
- "Do I like this Wine / Variety?" **computed** by aggregating Tastings.
- **Discovery**: nearest-neighbour over the enjoyed-grape set, each
  recommendation explained by the grape(s) it sits near.
- Multiple **Drinkers** with a plain switcher (no auth); basic management of
  Wines / Varieties / Companions.

**Deferred (additive, no re-architecture):**
- Contextual discovery (Note→facet re-indexing) and conversational exploration.
- Source-aware Provenance ladder / external dataset imports.
- Vintage as part of Wine identity.
- Authentication/authorization (login, passwords, admin).
