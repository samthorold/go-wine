-- Companions: personal-zone reference data — the name of a person a Drinker
-- drank with. Scoped to the Drinker who owns them (drinker_id), never linked to
-- a Drinker as an identity: a Companion is a name, even if that person also
-- happens to be a Drinker. The two never merge, which is what keeps the app free
-- of cross-Drinker sharing or consent.

CREATE TABLE IF NOT EXISTS companions (
    id         TEXT PRIMARY KEY,
    drinker_id TEXT NOT NULL REFERENCES drinkers(id),
    name       TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_companions_drinker ON companions (drinker_id, name);

-- A Tasting records who you were with: a many-to-many link between a Tasting and
-- the Companions present. ON DELETE CASCADE keeps the link table tidy if either
-- side is removed.
CREATE TABLE IF NOT EXISTS tasting_companions (
    tasting_id   TEXT NOT NULL REFERENCES tastings(id) ON DELETE CASCADE,
    companion_id TEXT NOT NULL REFERENCES companions(id) ON DELETE CASCADE,
    PRIMARY KEY (tasting_id, companion_id)
);

CREATE INDEX IF NOT EXISTS idx_tasting_companions_companion ON tasting_companions (companion_id);
