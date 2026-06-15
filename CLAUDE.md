# go-wine

A personal/household web app for logging the wine you drink, learning which
grapes you like, and discovering new ones.

- **Language & domain model:** [`CONTEXT.md`](./CONTEXT.md) — the ubiquitous
  language. Use these terms exactly (Variety, Wine, Style, Composition, Tasting,
  Note, Companion, Drinker, Taste profile, Discovery, Provenance).
- **System design (implementation-agnostic):** [`docs/system-design/`](./docs/system-design/)
  — the discovery mechanism, capture model, data-ownership partition, and the
  **v1 line** (see its `README.md`). Read before changing behaviour.

## Stack

- **Go**, server-rendered **HTMX**.
- Routing: stdlib `net/http` (Go 1.22+ `ServeMux`, method + path patterns). No
  web framework.
- Templating: **`templ`** (compile-time type-safe components; suits HTMX
  fragments). Build step via `templ generate`.
- Storage: **Postgres** via **`pgx`**. Schema via migrations.
- Dev/run: **docker-compose** — app + db as services. `pgvector` is the intended
  home for the deferred contextual/embedding discovery, so the db image carries
  it.

## Architecture — ports & adapters (onion)

- **Domain (core):** entities/value-objects (Variety, Wine, Composition, Tasting,
  Drinker, Companion); domain services (**Discovery**, **seed-merge**);
  repository **ports** (interfaces). No HTTP, no SQL.
- **Application (use-cases):** command handlers and query handlers. The CQRS
  split lives here.
- **Adapters (rim):** Postgres repositories, in-memory repositories, the
  HTTP+`templ` web adapter, the seed loader. HTMX and Postgres are *both* just
  adapters.

**One repository per aggregate root**, not per table: `WineRepository` owns
Wine+Composition; `VarietyRepository` owns Variety+characteristics+provenance;
plus `TastingRepository`, `DrinkerRepository`, `CompanionRepository`.

### DDD only where it earns its keep — and that line will move

The rich domain is small and concentrated; most of the app is thin CRUD. Apply
the full domain-model treatment **only** where there are real invariants or rich
logic, and keep the rest thin. **This placement is deliberate and expected to
change as behaviour accretes** — when a CRUD concept grows real rules, it gets
promoted into the domain.

Today's placement:

- **Command-side, through the domain model:** the **provenance seed-merge** (a
  re-seed never clobbers a confirmed value) and **Composition validity**
  (proportions sane, ≥1 Variety). Logging a Tasting is near-CRUD with a
  range-checked rating.
- **Query-side domain services (read-only, skip the aggregates):** **Discovery**,
  **Taste profile**, and "do I like this Wine/Variety" — computed projections
  over Tastings + characteristics. The richest logic in the app, and unit-tested
  in isolation against in-memory fixtures with no DB.
- **Thin CRUD (do *not* dress as aggregates):** managing Companions, browsing/
  editing Wines, Varieties, Tastings, reference fields.

**CQRS here is lightweight:** read services return view-model DTOs and may hit
Postgres directly; write paths go through the domain. *Not* event sourcing, *not*
two databases — one Postgres, one schema.

### Multi-tenancy, no auth (yet)

Multiple **Drinkers** exist; the app holds an **active Drinker** via a plain
switcher (profile-selection, not login). Scope every personal-zone query by
Drinker. There is no authentication/authorization yet and no security boundary —
an accepted property while it serves a household. See
[`docs/system-design/data-ownership.md`](./docs/system-design/data-ownership.md).

## Testing — TDD

- **Unit tests** run against **in-memory repository adapters** — fast, no
  container. This is what the repository ports buy us. The discovery/profile
  logic is exercised here against hand-built fixtures.
- **Integration tests** drive the whole app against **real Postgres** in the
  compose stack, each test wrapped in a **rolled-back transaction** for speed and
  isolation.
