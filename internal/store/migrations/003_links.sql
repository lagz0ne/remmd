CREATE TABLE IF NOT EXISTS links (
    id TEXT PRIMARY KEY,
    relationship_type TEXT NOT NULL,
    rationale_claim TEXT NOT NULL DEFAULT '',
    rationale_scope TEXT NOT NULL DEFAULT '',
    rationale_exclusions TEXT NOT NULL DEFAULT '',
    state TEXT NOT NULL DEFAULT 'pending',
    left_intervention TEXT NOT NULL DEFAULT 'notify',
    right_intervention TEXT NOT NULL DEFAULT 'notify',
    proposer_id TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS link_sections (
    link_id TEXT NOT NULL REFERENCES links(id),
    section_id TEXT NOT NULL REFERENCES sections(id),
    side TEXT NOT NULL CHECK(side IN ('left', 'right')),
    PRIMARY KEY(link_id, section_id, side)
);

CREATE TABLE IF NOT EXISTS thread_entries (
    id TEXT PRIMARY KEY,
    link_id TEXT NOT NULL REFERENCES links(id),
    entry_type TEXT NOT NULL CHECK(entry_type IN ('comment', 'system')),
    principal_id TEXT NOT NULL DEFAULT '',
    body TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_link_sections_section ON link_sections(section_id);

CREATE INDEX IF NOT EXISTS idx_thread_entries_link ON thread_entries(link_id, created_at);
