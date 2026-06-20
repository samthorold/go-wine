# Look & Feel

How the app is styled. Like [hypermedia-ui.md](./hypermedia-ui.md) this is a
posture, not a pixel spec: it fixes *where styling decisions live* and *how much
of the markup carries them*, so the choice survives a change of stylesheet.

## Minimal, classless, dark by default

The visual goal is **minimal**: legible semantic HTML with sane spacing and a
dark palette, not a designed brand. That goal picks the tool. A utility-class
framework (Bootstrap and friends) would buy comprehensive components at the cost
of sprinkling `class="..."` through every `templ` component — markup noise that
works *against* a household CRUD app whose templates are already plain `<form>`,
`<label>`, `<select>`, `<article>`. So the styling layer is **classless**: it
styles semantic elements directly, and the markup stays as close to bare HTML as
the interaction allows.

The concrete realisation is **[Pico.css](https://picocss.com) v2**, included as a
single CDN `<link>` in `Layout` alongside the htmx script. It is an adapter, in
the same sense htmx and Postgres are: the postures here would hold under any
classless stylesheet.

## The server owns the theme; there is no toggle

Dark is hard-coded — `data-theme="dark"` on `<html>` — with **no light styles
and no switcher**. This is the same reasoning that makes the app
[not JS-optional](./hypermedia-ui.md#the-interaction-runtime-is-required-the-app-is-not-js-optional):
the audience is a household on devices it controls, so a per-Drinker theme
preference is complexity bought for a need that doesn't exist. A toggle is
**deferred, additive, no re-architecture** — it would be a persisted preference
read into the same `data-theme` attribute the server already sets.

## Classes are the exception, custom CSS is rationed

Two escape hatches, both deliberately small:

- **A class appears only when the stylesheet requires one** — Pico's `.container`
  wrapper on `<main>`, its `<nav>`/`<ul>` chrome pattern, a grid when a layout
  genuinely needs it. The default is no class.
- **Custom CSS is for domain accents only**, not for structure or layout — those
  are Pico's job. Today that is a single rule (the Tasting `.rating` colour),
  living in the one `<style>` block in `Layout`. New custom rules should be rare,
  named for the domain thing they accent, and theme off Pico's `--pico-*`
  variables where they can, so they track the palette.

## No client-side UI components

Going classless also rules out a framework's JavaScript widgets — modals,
dropdowns, toasts, collapse. That is not a loss to absorb but a direct
consequence of the [hypermedia posture](./hypermedia-ui.md): those widgets carry
their own client-side state and need re-initialising on swapped-in content, which
fights htmx's model where the server returns the markup and htmx swaps it. Any
interactivity is a server-rendered fragment swapped by htmx, never a JS component
mounted on the client.

## Decision rules for a screen

The sections above fix *where styling lives*. These fix *how a screen is
composed* — a short set of rules, each earning its place by changing a decision
rather than describing taste, and each holding *within* the classless,
ration-the-custom-CSS constraint above rather than loosening it.

- **One primary action per view.** Pico renders every `<button>` as a filled
  accent, so left alone a screen has as many "primary" actions as it has buttons
  — and none of them reads as primary. The filled accent is reserved for the
  single action the view exists to perform (`Log tasting`). Chrome and admin
  actions take Pico's `secondary`/`outline` so they recede. If a view has no one
  obvious primary action, that is usually a sign it is doing two jobs.

- **Density follows frequency.** What a Drinker does every visit gets the
  prominence; what they do rarely recedes — onto its own page if it earns one.
  The primary task surface stays task-only and is not crowded by occasional
  management controls.

- **Admin is a noun, not a widget.** Because there are
  [no client-side components](#no-client-side-ui-components), managing things
  (Drinkers, Companions) is a server-rendered **page reached by navigation** — a
  REST noun like every other — not a panel bolted into the global chrome. The
  everyday *use* of a thing (switching the active Drinker) may live in the
  chrome; *managing* the set of them does not.

- **Readable measure over full bleed.** Forms and prose are constrained to a
  comfortable column even though the shell (`<main class="container">`) is wide.
  A full-width `<select>` or `<textarea>` is a long line to scan and a long
  pointer journey; the content column is narrower than the chrome.

- **Domain accents are consistent across read and write.** A thing the app gives
  a [custom accent](#classes-are-the-exception-custom-css-is-rationed) looks the
  same wherever it appears — a Rating reads as stars *and* is entered as stars,
  not stars in the list and a number `<select>` in the form. One accent, both
  directions.
