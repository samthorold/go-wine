-- Variety reference entity: a grape (Shiraz, Pinot Noir, Riesling) in the global
-- reference zone. This first slice carries only identity (a name); Variety
-- characteristics and provenance arrive in a later slice.

CREATE TABLE IF NOT EXISTS varieties (
    id   TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

-- Seed a coherent starter set of common grapes. Idempotent: the UNIQUE name plus
-- ON CONFLICT DO NOTHING makes a re-seed a no-op, never duplicating or clobbering.
-- IDs are random hex to match the application's ID format.
INSERT INTO varieties (id, name) VALUES
    (md5(random()::text || clock_timestamp()::text), 'Cabernet Sauvignon'),
    (md5(random()::text || clock_timestamp()::text), 'Merlot'),
    (md5(random()::text || clock_timestamp()::text), 'Pinot Noir'),
    (md5(random()::text || clock_timestamp()::text), 'Syrah'),
    (md5(random()::text || clock_timestamp()::text), 'Grenache'),
    (md5(random()::text || clock_timestamp()::text), 'Mourvèdre'),
    (md5(random()::text || clock_timestamp()::text), 'Tempranillo'),
    (md5(random()::text || clock_timestamp()::text), 'Sangiovese'),
    (md5(random()::text || clock_timestamp()::text), 'Nebbiolo'),
    (md5(random()::text || clock_timestamp()::text), 'Malbec'),
    (md5(random()::text || clock_timestamp()::text), 'Chardonnay'),
    (md5(random()::text || clock_timestamp()::text), 'Sauvignon Blanc'),
    (md5(random()::text || clock_timestamp()::text), 'Riesling'),
    (md5(random()::text || clock_timestamp()::text), 'Pinot Grigio'),
    (md5(random()::text || clock_timestamp()::text), 'Chenin Blanc')
ON CONFLICT (name) DO NOTHING;
