-- C1: Document Types
ALTER TABLE documents ADD COLUMN doc_type TEXT NOT NULL DEFAULT '';

-- C4: Document Hierarchy
ALTER TABLE documents ADD COLUMN parent_doc_id TEXT REFERENCES documents(id);
CREATE INDEX IF NOT EXISTS idx_documents_parent ON documents(parent_doc_id);

-- C2: Section Kinds (semantic classification, separate from structural 'type')
ALTER TABLE sections ADD COLUMN kind TEXT NOT NULL DEFAULT '';

-- C3: Schema Templates (required section kinds per document type)
CREATE TABLE IF NOT EXISTS schema_templates (
    doc_type TEXT NOT NULL,
    required_kind TEXT NOT NULL,
    min_count INTEGER NOT NULL DEFAULT 1,
    PRIMARY KEY(doc_type, required_kind)
);

-- C8: Structural Relations (lightweight doc-to-doc edges, distinct from trust-bearing Links)
CREATE TABLE IF NOT EXISTS relations (
    id TEXT PRIMARY KEY,
    from_doc_id TEXT NOT NULL REFERENCES documents(id),
    to_doc_id TEXT NOT NULL REFERENCES documents(id),
    relation_type TEXT NOT NULL,
    metadata TEXT NOT NULL DEFAULT '{}',
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_relations_from ON relations(from_doc_id);
CREATE INDEX IF NOT EXISTS idx_relations_to ON relations(to_doc_id);

-- VF1: Missing index on sections.ref (hot path for FindSectionByRefGlobal)
CREATE INDEX IF NOT EXISTS idx_sections_ref ON sections(ref);

-- VF3: Per-version provenance metadata
ALTER TABLE section_versions ADD COLUMN metadata TEXT NOT NULL DEFAULT '{}';

-- C6: Full-Text Search (FTS5, content-sync mode with triggers)
CREATE VIRTUAL TABLE IF NOT EXISTS sections_fts USING fts5(
    title,
    content,
    kind,
    content=sections,
    content_rowid=rowid
);

-- Populate FTS from existing data
INSERT INTO sections_fts(rowid, title, content, kind)
SELECT rowid, title, content, '' FROM sections;

-- Keep FTS in sync via triggers
CREATE TRIGGER IF NOT EXISTS sections_fts_insert AFTER INSERT ON sections BEGIN
    INSERT INTO sections_fts(rowid, title, content, kind)
    VALUES (new.rowid, new.title, new.content, new.kind);
END;

CREATE TRIGGER IF NOT EXISTS sections_fts_update AFTER UPDATE ON sections BEGIN
    INSERT INTO sections_fts(sections_fts, rowid, title, content, kind)
    VALUES ('delete', old.rowid, old.title, old.content, old.kind);
    INSERT INTO sections_fts(rowid, title, content, kind)
    VALUES (new.rowid, new.title, new.content, new.kind);
END;

CREATE TRIGGER IF NOT EXISTS sections_fts_delete AFTER DELETE ON sections BEGIN
    INSERT INTO sections_fts(sections_fts, rowid, title, content, kind)
    VALUES ('delete', old.rowid, old.title, old.content, old.kind);
END;

-- VF7: Drop dead event store table
DROP TABLE IF EXISTS events;
