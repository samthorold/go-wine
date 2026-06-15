# go-wine

A personal/household web app for logging the wine you drink, learning which
grapes you like, and discovering new ones. Go + HTMX, server-rendered.

- **What it models and why:** [`CONTEXT.md`](./CONTEXT.md) (language) and
  [`docs/system-design/`](./docs/system-design/) (the design and its rationale).
- **How the code is built:** [`CLAUDE.md`](./CLAUDE.md) (stack, architecture,
  testing).

## Run it

Full stack (app + Postgres) in Docker:

```sh
make up                 # http://localhost:8080
```

Or locally with no database (in-memory store, seeded in process):

```sh
make tools              # once, installs templ
make run                # http://localhost:8080
```

Tests:

```sh
make test
```

> Editing a `*.templ` file? Run `make generate` (the `test`/`run`/`build`
> targets do this for you).

## Status

First vertical slice: pick an active **Drinker** (plain switcher, no login),
**log a Tasting** against a seeded **Wine** with a 5-point rating and a free-text
Note, and see your tastings — wired end-to-end through ports & adapters, on both
the in-memory and Postgres stores. See the deferred items in
[`docs/system-design/README.md`](./docs/system-design/README.md).
