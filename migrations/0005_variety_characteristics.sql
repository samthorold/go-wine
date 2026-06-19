-- Variety characteristics: intrinsic, conventional reference data about a grape
-- — the scalar axes (body, tannin, acidity, sweetness, alcohol) on a fixed 1..5
-- rubric, the typical flavour-note tags, and a single binary Provenance for the
-- bundle. Characteristics are a value object on the Variety aggregate (one
-- repository owns Variety + Characteristics), so the scalar axes live in a
-- 1:1 table keyed by the Variety and the flavour-note tags in a child table.
--
-- Provenance is 'default' (a neutral conventional seed) or 'confirmed' (vetted
-- by the Drinker). The no-clobber-on-reseed rule lives in the domain seed-merge,
-- NOT in a SQL ON CONFLICT: a re-seed reads the stored bundle, merges it through
-- the domain (which preserves any confirmed value), and writes the result back.

CREATE TABLE IF NOT EXISTS variety_characteristics (
    variety_id TEXT PRIMARY KEY REFERENCES varieties(id) ON DELETE CASCADE,
    body       INT  NOT NULL CHECK (body       BETWEEN 1 AND 5),
    tannin     INT  NOT NULL CHECK (tannin     BETWEEN 1 AND 5),
    acidity    INT  NOT NULL CHECK (acidity    BETWEEN 1 AND 5),
    sweetness  INT  NOT NULL CHECK (sweetness  BETWEEN 1 AND 5),
    alcohol    INT  NOT NULL CHECK (alcohol    BETWEEN 1 AND 5),
    provenance TEXT NOT NULL DEFAULT 'default' CHECK (provenance IN ('default', 'confirmed'))
);

CREATE TABLE IF NOT EXISTS variety_notes (
    variety_id TEXT NOT NULL REFERENCES varieties(id) ON DELETE CASCADE,
    note       TEXT NOT NULL,
    PRIMARY KEY (variety_id, note)
);
