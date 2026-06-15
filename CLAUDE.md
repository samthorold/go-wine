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

## HTMX & HTML conventions

The *why* — hypermedia-driven, required (not JS-optional), boosted-MPA, carved
along swap boundaries, server-authoritative validation — lives in
[`docs/system-design/hypermedia-ui.md`](./docs/system-design/hypermedia-ui.md).
These are the mechanics that encode it.

### Responses: pages vs fragments

- **One URL → one response shape.** An endpoint returns *either* a full page
  *or* a fragment, never both by sniffing `HX-Request`. A boosted page handler
  is oblivious to whether it was boosted.
- **Page** = a `GET` on a resource noun returns `Layout` + content
  (`TastingsPage`). Reached by boosted navigation.
- **Fragment** = a mutation or partial-update endpoint returns bare markup (no
  `Layout`); the caller owns `hx-target`/`hx-swap`.
- The line: a view a Drinker can navigate to and bookmark is a **page**; a piece
  that only ever fills a target inside an already-loaded page is a **fragment**.

### Routing

- **Pages are REST nouns**: `GET /tastings`, `GET /wines`, `GET /wines/{id}`.
- **Mutations use explicit verbs on the resource**: `POST /tastings`,
  `PUT /wines/{id}`, `DELETE /tastings/{id}` — not `POST /x/delete`.
- **Fragment GETs are sub-resources** named for what they fill:
  `GET /wines/{id}/edit`, `GET /tastings/validate`. **No `/partials/` marker
  prefix.**
- **No safe-method mutations.** A `GET` never mutates. A mutation that should
  land on a page returns **`303 See Other`** (e.g. `POST /switch` → `/tastings`).

### `templ` components

- `Layout` is the shell: `<html>`/head/`<body hx-boost="true">` + global chrome.
  It owns the htmx script include and the one-time `htmx.config.responseHandling`
  setup. Takes `children...`.
- `XxxPage` composes `Layout` + the page's content.
- **A component exists for each page, each fragment a request swaps, or anything
  reused — inline the rest.** Don't pre-carve a component tree.
- A swap target **owns its `id` and its own `hx-*` attributes**, and is rendered
  by one component so first-paint and re-render can't drift.
- **Naming**: `Layout`; `XxxPage`; region components named for the region
  (`LogForm`, `TastingList`); items singular (`TastingRow`).
- **Export = a handler renders it directly** (`TastingsPage`, `LogForm`,
  `TastingList`); compositional chrome stays unexported (`drinkerSwitcher`).
- Components render **view-model DTOs, never domain entities**: query-handler
  outputs (`app.TastingView`) or `views`-package shapes built in the adapter
  (`DrinkerOption`, `WineOption`).

### Swaps & OOB

- **Always set `hx-target` and `hx-swap` explicitly.** Never rely on the default
  (target = the triggering element).
- **Default is a wholesale `innerHTML` re-render of the target region.** Reach
  for `afterbegin`/`beforeend` (append one item) only as an optimization;
  lists here are small enough that you never need to.
- **OOB (`hx-swap-oob`) only for genuinely disconnected regions** updated by one
  response — after first asking "can I just widen the target?".
- **Mutation response = the primary target (the region that changes in *every*
  outcome) + one OOB fragment per *other* disconnected region it changed.**
  Logging a Tasting: `hx-target="#log-form"`, `hx-swap="outerHTML"`; success
  returns a fresh empty `LogForm` **+ `TastingList` OOB** (list updates,
  empty-state clears); failure returns `LogForm` with errors (list untouched).

### Validation & status codes

- **HTML5 constraints** (`required`, `type`, `min`/`max`) are the first line — a
  courtesy that blocks the obvious before a request is sent.
- **The domain command handler is the authority.** Invariants live in the
  domain, never in the form.
- **Status codes** (htmx swaps on 2xx only): configure
  `htmx.config.responseHandling` once in `Layout` to also swap **`422`**. Then:
  `200` success; **`422`** validation failure (re-render the form); `303`
  navigational mutation; other `4xx`/`5xx` no swap.
- **Failure re-renders the form** with entered values preserved: **inline errors
  in each field's slot** for field-attributable failures, a **form-level banner**
  for non-field failures. The form component therefore takes a view-model
  carrying options + entered values + a `field → message` error map (empty on
  first paint), not just its options.

## Testing — TDD

- **Unit tests** run against **in-memory repository adapters** — fast, no
  container. This is what the repository ports buy us. The discovery/profile
  logic is exercised here against hand-built fixtures.
- **Integration tests** drive the whole app against **real Postgres** in the
  compose stack, each test wrapped in a **rolled-back transaction** for speed and
  isolation.
