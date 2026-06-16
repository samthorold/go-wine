-- Wine Composition: the Varieties that make up a Wine, with rough proportions.
-- The Composition is a value object on the Wine aggregate — one repository owns
-- Wine + Composition — so this is a link table keyed by the Wine, not an entity
-- with its own identity.
--
-- proportion is an integer percentage (1..100). The aggregate invariant (≥1
-- Variety, proportions summing to ~100% within tolerance) is enforced in the
-- domain; the column-level CHECK here only guards the per-row range, since the
-- cross-row sum-to-100% rule isn't naturally a column constraint.

CREATE TABLE IF NOT EXISTS wine_varieties (
    wine_id    TEXT NOT NULL REFERENCES wines(id) ON DELETE CASCADE,
    variety_id TEXT NOT NULL REFERENCES varieties(id) ON DELETE CASCADE,
    proportion INT  NOT NULL CHECK (proportion BETWEEN 1 AND 100),
    PRIMARY KEY (wine_id, variety_id)
);

CREATE INDEX IF NOT EXISTS idx_wine_varieties_variety ON wine_varieties (variety_id);
