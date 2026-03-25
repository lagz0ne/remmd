package store_test

import (
	"testing"

	"github.com/lagz0ne/remmd/internal/store"
)

func TestMigrate_CreatesSchemaVersionTable(t *testing.T) {
	t.Parallel()
	db, err := store.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("OpenDB error: %v", err)
	}
	defer store.CloseDB(db)

	if err := store.Migrate(db); err != nil {
		t.Fatalf("Migrate error: %v", err)
	}

	// schema_version table should exist
	var count int
	err = db.QueryRow("SELECT count(*) FROM schema_version").Scan(&count)
	if err != nil {
		t.Fatalf("schema_version table does not exist: %v", err)
	}
}

func TestMigrate_Idempotent(t *testing.T) {
	t.Parallel()
	db, err := store.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("OpenDB error: %v", err)
	}
	defer store.CloseDB(db)

	if err := store.Migrate(db); err != nil {
		t.Fatalf("first Migrate error: %v", err)
	}
	if err := store.Migrate(db); err != nil {
		t.Fatalf("second Migrate error: %v", err)
	}
}

func TestMigrate_EventsTableDropped(t *testing.T) {
	t.Parallel()
	db, err := store.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("OpenDB error: %v", err)
	}
	defer store.CloseDB(db)

	if err := store.Migrate(db); err != nil {
		t.Fatalf("Migrate error: %v", err)
	}

	// After migration 007, the events table should NOT exist
	_, err = db.Exec("SELECT 1 FROM events")
	if err == nil {
		t.Fatal("expected error querying dropped events table, but got nil")
	}
}

func TestMigrate_SchemaVersionTracked(t *testing.T) {
	t.Parallel()
	db, err := store.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("OpenDB error: %v", err)
	}
	defer store.CloseDB(db)

	if err := store.Migrate(db); err != nil {
		t.Fatalf("Migrate error: %v", err)
	}

	var version int
	err = db.QueryRow("SELECT MAX(version) FROM schema_version").Scan(&version)
	if err != nil {
		t.Fatalf("query schema_version failed: %v", err)
	}
	if version < 1 {
		t.Fatalf("expected schema version >= 1, got %d", version)
	}
}

func TestMigration007_Capabilities(t *testing.T) {
	t.Parallel()
	db, err := store.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("OpenDB error: %v", err)
	}
	defer store.CloseDB(db)

	if err := store.Migrate(db); err != nil {
		t.Fatalf("Migrate error: %v", err)
	}

	// documents table has doc_type and parent_doc_id columns
	_, err = db.Exec(`INSERT INTO documents (id, title, owner_id, doc_type, parent_doc_id)
		VALUES ('doc-1', 'Root', 'user-1', 'spec', NULL)`)
	if err != nil {
		t.Fatalf("INSERT doc with doc_type failed: %v", err)
	}
	_, err = db.Exec(`INSERT INTO documents (id, title, owner_id, doc_type, parent_doc_id)
		VALUES ('doc-2', 'Child', 'user-1', 'spec', 'doc-1')`)
	if err != nil {
		t.Fatalf("INSERT child doc with parent_doc_id failed: %v", err)
	}
	var docType, parentDocID string
	err = db.QueryRow("SELECT doc_type, parent_doc_id FROM documents WHERE id = 'doc-2'").
		Scan(&docType, &parentDocID)
	if err != nil {
		t.Fatalf("SELECT doc_type/parent_doc_id failed: %v", err)
	}
	if docType != "spec" {
		t.Errorf("doc_type = %q, want %q", docType, "spec")
	}
	if parentDocID != "doc-1" {
		t.Errorf("parent_doc_id = %q, want %q", parentDocID, "doc-1")
	}

	// sections table has kind column
	_, err = db.Exec(`INSERT INTO sections (id, doc_id, ref, type, kind)
		VALUES ('sec-1', 'doc-1', '@s1', 'paragraph', 'requirement')`)
	if err != nil {
		t.Fatalf("INSERT section with kind failed: %v", err)
	}
	var kind string
	err = db.QueryRow("SELECT kind FROM sections WHERE id = 'sec-1'").Scan(&kind)
	if err != nil {
		t.Fatalf("SELECT kind failed: %v", err)
	}
	if kind != "requirement" {
		t.Errorf("kind = %q, want %q", kind, "requirement")
	}

	// section_versions has metadata column
	_, err = db.Exec(`INSERT INTO section_versions (id, section_id, version, content, content_hash, metadata)
		VALUES ('sv-1', 'sec-1', 1, 'hello', 'abc', '{"author":"bot"}')`)
	if err != nil {
		t.Fatalf("INSERT section_version with metadata failed: %v", err)
	}
	var meta string
	err = db.QueryRow("SELECT metadata FROM section_versions WHERE id = 'sv-1'").Scan(&meta)
	if err != nil {
		t.Fatalf("SELECT metadata failed: %v", err)
	}
	if meta != `{"author":"bot"}` {
		t.Errorf("metadata = %q, want %q", meta, `{"author":"bot"}`)
	}

	// schema_templates table exists (INSERT + SELECT)
	_, err = db.Exec(`INSERT INTO schema_templates (doc_type, required_kind, min_count)
		VALUES ('spec', 'requirement', 2)`)
	if err != nil {
		t.Fatalf("INSERT schema_templates failed: %v", err)
	}
	var minCount int
	err = db.QueryRow("SELECT min_count FROM schema_templates WHERE doc_type = 'spec' AND required_kind = 'requirement'").
		Scan(&minCount)
	if err != nil {
		t.Fatalf("SELECT schema_templates failed: %v", err)
	}
	if minCount != 2 {
		t.Errorf("min_count = %d, want 2", minCount)
	}

	// relations table exists (INSERT + SELECT)
	_, err = db.Exec(`INSERT INTO relations (id, from_doc_id, to_doc_id, relation_type)
		VALUES ('rel-1', 'doc-1', 'doc-2', 'implements')`)
	if err != nil {
		t.Fatalf("INSERT relations failed: %v", err)
	}
	var relType string
	err = db.QueryRow("SELECT relation_type FROM relations WHERE id = 'rel-1'").Scan(&relType)
	if err != nil {
		t.Fatalf("SELECT relations failed: %v", err)
	}
	if relType != "implements" {
		t.Errorf("relation_type = %q, want %q", relType, "implements")
	}

	// FTS5 works: the trigger should have auto-inserted sec-1 into sections_fts
	// Also test a direct match
	_, err = db.Exec(`INSERT INTO sections (id, doc_id, ref, type, kind, title, content)
		VALUES ('sec-2', 'doc-1', '@s2', 'paragraph', 'note', 'Test Title', 'searchable content here')`)
	if err != nil {
		t.Fatalf("INSERT section for FTS test failed: %v", err)
	}
	var ftsRowid int
	err = db.QueryRow("SELECT rowid FROM sections_fts WHERE sections_fts MATCH 'searchable'").Scan(&ftsRowid)
	if err != nil {
		t.Fatalf("FTS5 MATCH query failed: %v", err)
	}
	if ftsRowid == 0 {
		t.Error("FTS5 returned rowid 0, expected non-zero")
	}

	// events table does NOT exist
	_, err = db.Exec("SELECT 1 FROM events")
	if err == nil {
		t.Fatal("expected error querying dropped events table, but got nil")
	}

	// idx_sections_ref index exists
	var indexCount int
	err = db.QueryRow("SELECT count(*) FROM sqlite_master WHERE type='index' AND name='idx_sections_ref'").
		Scan(&indexCount)
	if err != nil {
		t.Fatalf("query sqlite_master for idx_sections_ref failed: %v", err)
	}
	if indexCount != 1 {
		t.Errorf("idx_sections_ref count = %d, want 1", indexCount)
	}
}
