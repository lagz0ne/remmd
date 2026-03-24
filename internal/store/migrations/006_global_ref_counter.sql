CREATE TABLE IF NOT EXISTS ref_counter (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    next_seq INTEGER NOT NULL DEFAULT 1
);

-- Seed from existing sections: start after the total count of existing sections
INSERT INTO ref_counter (id, next_seq)
SELECT 1, COALESCE((SELECT COUNT(*) + 1 FROM sections), 1)
WHERE NOT EXISTS (SELECT 1 FROM ref_counter);
