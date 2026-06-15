-- First vertical slice: the tables needed to log a Tasting end-to-end.
-- Reference zone (global): drinkers (the personal-zone owners) and wines.
-- Personal zone: tastings, scoped by drinker_id.
-- Composition, Companions, and Variety characteristics arrive in later slices.

CREATE TABLE IF NOT EXISTS drinkers (
    id   TEXT PRIMARY KEY,
    name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS wines (
    id       TEXT PRIMARY KEY,
    producer TEXT NOT NULL,
    name     TEXT NOT NULL,
    style    TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS tastings (
    id         TEXT PRIMARY KEY,
    drinker_id TEXT NOT NULL REFERENCES drinkers(id),
    wine_id    TEXT NOT NULL REFERENCES wines(id),
    vintage    INT,
    rating     INT NOT NULL CHECK (rating BETWEEN 1 AND 5),
    note       TEXT NOT NULL DEFAULT '',
    drunk_on   TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_tastings_drinker ON tastings (drinker_id, drunk_on DESC);
