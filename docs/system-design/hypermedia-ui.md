# The Hypermedia UI

How a Drinker interacts with the app. This is a design posture, not a styling
choice: it fixes where application state lives, what a server response *is*, and
how a failed command reaches the screen. HTMX and `templ` are the adapters that
realise it; the postures below would hold under any equivalent hypermedia
runtime.

## The server is the single source of truth for markup and state

The app is **hypermedia-driven**: every screen is HTML the server renders, and
every change to what a Drinker sees is the server returning new HTML to swap in.
There is no client-side application state, no JSON data API, and no
model-on-the-client to keep in sync. The interaction runtime issues requests and
swaps responses; it does not hold the domain.

This shape is forced by where the value of the app lives. The hard logic —
Discovery, the Taste profile, "do I like this Wine?" — is **computed server-side
by aggregating Tastings** against seeded characteristics. None of it can run on
the client without shipping the data and the math there. A JSON API with a
client that re-derives views would therefore duplicate the projection logic on
both sides and invent a second source of truth for state that only the server
can authoritatively compute. Returning rendered HTML keeps the one source of
truth where the data and the logic already are.

## The interaction runtime is required; the app is not JS-optional

The app **requires** its hypermedia runtime to function; core flows are not built
to also work as plain no-JS form posts. This is not the usual progressive-
enhancement default, and the reason it inverts here is the deployment: this is a
personal/household app on devices the household controls, so the "Drinker with
scripting disabled" that progressive enhancement protects does not exist. The
friction constraint that governs capture (see
[capturing-tastings.md](./capturing-tastings.md)) is about *effort to log a
Tasting*, not about browser support — it argues for fewer taps, not for a no-JS
fallback. Building every command path to render two ways (a swapped fragment and
a full redirected page) would be ongoing complexity bought for an audience of
nobody.

## Navigation is full pages; interactivity is partial swaps of them

The app is a **multi-page application whose links and forms are boosted**: moving
between resources (the Tastings list, a Wine, the Discovery view) loads a **full
server-rendered page**, and the runtime swaps it in while keeping the URL real
and bookmarkable. Within a loaded page, an interaction that changes only part of
it — logging a Tasting, an inline edit — returns just **that part**.

So a response is one of two things: a **page** (a complete representation of a
resource, reached by navigation) or a **fragment** (a piece of a page that one
interaction replaces). The dividing line is whether a Drinker can navigate to and
bookmark it: a navigable view is a page; a piece that only ever fills a target
inside an already-loaded page is a fragment. A given endpoint returns one of
these, never both depending on who is asking — which is what lets a page handler
stay oblivious to whether the request was boosted.

The tempting alternative is a single-shell app that loads once and thereafter
only swaps fragments. It is rejected because it throws away the property that
makes a CRUD-shaped household app cheap to hold: every resource staying an
independently addressable, server-rendered URL. The shell approach recreates
routing, history, and addressability in the client — the very client-side
machinery the first posture exists to avoid.

### The unit of structure is the swap, not the component tree

Because partial updates return pieces of pages, the page is divided **along the
boundaries of what an interaction replaces** — not into a tidy tree of nested
components for its own sake. A piece of a page earns its own existence as a unit
when some interaction swaps it independently, or when it is reused across pages.
Everything else stays inline in its page.

This is a deliberate rejection of component-decomposition-as-default (the habit
imported from client-side frameworks, where the tree exists to drive
reconciliation). Here there is no reconciliation and no client state, so a unit
that is never swapped and never reused buys nothing but indirection. The question
that creates a boundary is "does a request replace this on its own?", and the
answer is a property of the *interactions*, not of how neatly the markup nests.

## A failed command renders by re-rendering its form

When a command is invalid — a rating out of range, a Composition that does not
hold — the **domain** is the authority that rejects it, and the failure reaches
the Drinker as the **same form, re-rendered with the entered values preserved and
the errors shown against the offending fields**. Validation is not a separate
client-side concept that can disagree with the domain: client-level input
constraints are a courtesy that catches the obvious before a request is sent, but
the server's domain check is the gate, and its rejection is rendered HTML like
everything else.

This follows directly from the first posture. If the only source of truth for
state is the server, then the truth about *why a command failed* is also the
server's, and the honest way to surface it is to send back the corrected view of
the form. A client that knew the validation rules well enough to render errors
without asking would be re-implementing the domain's invariants on the client —
reintroducing the second source of truth the design exists to avoid, and one that
silently drifts from the real rules as they evolve.
