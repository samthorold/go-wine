# Data Ownership

The data model splits into two zones, and keeping them apart is what makes
multi-tenancy cheap and authentication a wholly separate, deferred concern.

## Two zones

**Personal zone** — owned by a **Drinker** and siloed per Drinker: **Tastings**,
their **ratings**, **Companions**, and the derived **Taste profile**. This is "a
drinker's history and palate."

**Reference zone** — global, objective, and identical for every drinker:
**Variety** characteristics, **Style** → **Composition** defaults, and the
**Wine** catalogue. This is "facts about the world." It is seeded from
conventional knowledge and corrected *toward a single shared truth*, so its
**Provenance** stays global — *default guess* vs *human-verified fact* — never
split per drinker. There is no per-user version of a grape's tannin any more than
there is a per-user boiling point.

A **Companion** sits firmly in the personal zone: it is the *name of a person I
drank with*, not a user of the system. The two never merge — even a companion who
happened to use the app would still be, to me, a name attached to my Tastings.

## Why this makes multi-drinker additive, not a rebuild

Multi-drinker systems are usually expensive because of two things this design
does not have:

- **Sharing/consent between users** — absent, because Companions are names, not
  identities. Nobody's personal data is ever exposed to another drinker, so there
  is no friend graph, invitation flow, or permission model to build.
- **Contested per-user data** — absent, because the reference zone is objective
  truth, not opinion. A correction to Nebbiolo's tannin improves the one shared
  fact for everyone; it is not a personal override that must be reconciled.

So multi-tenancy is purely additive: every personal-zone entity carries a
**Drinker** owner; the reference zone is untouched; and no social subsystem comes
into existence. The hard parts of multi-user are absent *by construction* of
these two facts, not deferred by luck.

## Multi-tenancy without authentication

Multiple **Drinkers** exist from the start, but there is no login, no password,
no admin. The app holds an **active Drinker** chosen through a plain switcher —
profile-selection, not sign-in — and scopes the personal zone to it.

This works because two questions usually fused are in fact orthogonal:

- *Whose data is this?* — multi-tenancy, a modelling question, answered by the
  **Drinker** owner on every personal-zone entity.
- *Are you allowed to be this Drinker?* — authentication/authorization, a
  security question about keeping someone *out*.

The data model needs only the first. Building the switcher now exercises every
multi-Drinker path for real, so the partition is tested rather than theoretical,
while the entire security surface — sessions, passwords, admin — stays deferred
until there is actually someone to exclude. There is no security boundary today:
anyone using the running app can switch to any Drinker. That is an accepted
property while the app serves a household, not a defect to patch.
