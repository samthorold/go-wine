-- Style → Composition seed: the input-side body of conventional knowledge that
-- mirrors the variety-characteristics rubric on the output side. Two concerns:
--
-- 1. A Wine's Composition now carries a binary Provenance, exactly as a Variety's
--    Characteristics does: 'default' when filled from the Style → Composition seed
--    (a conventional guess), 'confirmed' once the Drinker names or edits the grapes
--    themselves. The Composition is a value object on the Wine aggregate (one
--    repository owns Wine + Composition), so the flag lives per-Wine. Earlier
--    wine_varieties rows are all Drinker-or-seed data with no recorded provenance;
--    the column defaults to 'default', the safe (overwritable) value.
--
--    As with characteristics, the no-clobber-on-reseed rule lives in the domain
--    seed-merge (MergeComposition), NOT in a SQL ON CONFLICT: a re-seed reads the
--    stored Composition, merges it through the domain (preserving any confirmed
--    value), and writes the result back. Provenance is per-Wine, so it sits on the
--    wines row rather than repeated across every wine_varieties row.
--
-- 2. style_compositions holds the reference map itself — a Style name → default
--    blend over Varieties, joined to varieties BY NAME (the stable key, since each
--    store mints its own random Variety IDs). It is reference data the seed-merge
--    reads; the authoritative copy lives in code (internal/seed/styles.go) and is
--    shared with the in-memory store, mirroring how the characteristics rubric is
--    wired into both stores.

ALTER TABLE wines
    ADD COLUMN IF NOT EXISTS composition_provenance TEXT NOT NULL DEFAULT 'default'
    CHECK (composition_provenance IN ('default', 'confirmed'));

CREATE TABLE IF NOT EXISTS style_compositions (
    style       TEXT NOT NULL,
    variety_name TEXT NOT NULL,
    proportion  INT  NOT NULL CHECK (proportion BETWEEN 1 AND 100),
    PRIMARY KEY (style, variety_name)
);

INSERT INTO style_compositions (style, variety_name, proportion) VALUES
    ('Chianti',         'Sangiovese',          85),
    ('Chianti',         'Merlot',              15),
    ('GSM',             'Grenache',            50),
    ('GSM',             'Syrah',               30),
    ('GSM',             'Mourvèdre',           20),
    ('Chablis',         'Chardonnay',         100),
    ('Bordeaux Blend',  'Cabernet Sauvignon',  70),
    ('Bordeaux Blend',  'Merlot',              30),
    ('Rioja',           'Tempranillo',         90),
    ('Rioja',           'Grenache',            10),
    ('Côtes du Rhône',  'Grenache',            60),
    ('Côtes du Rhône',  'Syrah',               40)
ON CONFLICT (style, variety_name) DO NOTHING;
