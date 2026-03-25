package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lagz0ne/remmd/internal/core"
)

// DocumentRepo is the SQLite implementation of core.DocumentRepository.
type DocumentRepo struct {
	db *sql.DB
}

// NewDocumentRepo creates a DocumentRepo backed by the given *sql.DB.
func NewDocumentRepo(db *sql.DB) *DocumentRepo {
	return &DocumentRepo{db: db}
}

func (r *DocumentRepo) CreateDocument(ctx context.Context, doc *core.Document) error {
	if doc.CreatedAt.IsZero() {
		doc.CreatedAt = time.Now()
	}
	var parentDocID *string
	if doc.ParentDocID != "" {
		parentDocID = &doc.ParentDocID
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO documents (id, title, owner_id, status, source, doc_type, parent_doc_id, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		doc.ID, doc.Title, doc.OwnerID, doc.Status, doc.Source, doc.DocType, parentDocID,
		formatTime(doc.CreatedAt),
	)
	if err != nil {
		return fmt.Errorf("insert document: %w", err)
	}
	return nil
}

func (r *DocumentRepo) FindDocumentByID(ctx context.Context, id string) (*core.Document, error) {
	var doc core.Document
	var createdAt string
	var parentDocID sql.NullString
	err := r.db.QueryRowContext(ctx,
		`SELECT id, title, owner_id, status, source, doc_type, parent_doc_id, created_at FROM documents WHERE id = ?`, id,
	).Scan(&doc.ID, &doc.Title, &doc.OwnerID, &doc.Status, &doc.Source, &doc.DocType, &parentDocID, &createdAt)
	if err == sql.ErrNoRows {
		return nil, core.ErrNotFound{Entity: "document", ID: id}
	}
	if err != nil {
		return nil, fmt.Errorf("find document %s: %w", id, err)
	}
	if parentDocID.Valid {
		doc.ParentDocID = parentDocID.String
	}
	doc.CreatedAt = parseTime(createdAt)
	return &doc, nil
}

func (r *DocumentRepo) ListDocuments(ctx context.Context) ([]*core.Document, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, title, owner_id, status, source, doc_type, parent_doc_id, created_at FROM documents ORDER BY created_at`)
	if err != nil {
		return nil, fmt.Errorf("list documents: %w", err)
	}
	defer rows.Close()

	var docs []*core.Document
	for rows.Next() {
		var doc core.Document
		var createdAt string
		var parentDocID sql.NullString
		if err := rows.Scan(&doc.ID, &doc.Title, &doc.OwnerID, &doc.Status, &doc.Source, &doc.DocType, &parentDocID, &createdAt); err != nil {
			return nil, fmt.Errorf("scan document: %w", err)
		}
		if parentDocID.Valid {
			doc.ParentDocID = parentDocID.String
		}
		doc.CreatedAt = parseTime(createdAt)
		docs = append(docs, &doc)
	}
	return docs, rows.Err()
}

// populateSection sets parsed fields on a Section from raw scanned strings.
func populateSection(s *core.Section, refStr, typStr, kind string, parentRefStr sql.NullString, contentType, metadata string) {
	ref, _ := core.ParseRef(refStr)
	s.Ref = ref
	s.Type = core.SectionType(typStr)
	s.Kind = kind
	if parentRefStr.Valid {
		pref, _ := core.ParseRef(parentRefStr.String)
		s.ParentRef = &pref
	}
	s.ContentType = core.ContentType(contentType)
	s.Metadata = metadata
}

func (r *DocumentRepo) CreateSection(ctx context.Context, s *core.Section) error {
	var parentID *string
	if s.ParentRef != nil {
		var pid string
		err := r.db.QueryRowContext(ctx,
			`SELECT id FROM sections WHERE doc_id = ? AND ref = ?`,
			s.DocID, s.ParentRef.String(),
		).Scan(&pid)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("lookup parent ref: %w", err)
		}
		if err == nil {
			parentID = &pid
		}
	}

	contentType := string(s.ContentType)
	if contentType == "" {
		contentType = string(core.ContentNative)
	}
	metadata := s.Metadata
	if metadata == "" {
		metadata = "{}"
	}

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO sections (id, doc_id, ref, type, kind, title, content, content_hash, parent_id, "order", content_type, metadata)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		s.ID, s.DocID, s.Ref.String(), string(s.Type), s.Kind, s.Title, s.Content, s.ContentHash, parentID, s.Order, contentType, metadata,
	)
	if err != nil {
		return fmt.Errorf("insert section: %w", err)
	}
	return nil
}

func (r *DocumentRepo) ListSections(ctx context.Context, docID string) ([]*core.Section, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT s.id, s.doc_id, s.ref, s.type, s.kind, s.title, s.content, s.content_hash, p.ref, s."order", s.content_type, s.metadata
		 FROM sections s
		 LEFT JOIN sections p ON s.parent_id = p.id
		 WHERE s.doc_id = ?
		 ORDER BY s."order"`, docID,
	)
	if err != nil {
		return nil, fmt.Errorf("list sections: %w", err)
	}
	defer rows.Close()

	var sections []*core.Section
	for rows.Next() {
		var s core.Section
		var refStr, typStr, kind string
		var parentRefStr sql.NullString
		var contentType, metadata string
		if err := rows.Scan(&s.ID, &s.DocID, &refStr, &typStr, &kind, &s.Title, &s.Content, &s.ContentHash, &parentRefStr, &s.Order, &contentType, &metadata); err != nil {
			return nil, fmt.Errorf("scan section: %w", err)
		}
		populateSection(&s, refStr, typStr, kind, parentRefStr, contentType, metadata)
		sections = append(sections, &s)
	}
	return sections, rows.Err()
}

func (r *DocumentRepo) FindSectionByRef(ctx context.Context, docID string, ref string) (*core.Section, error) {
	var s core.Section
	var refStr, typStr, kind string
	var parentRefStr sql.NullString
	var contentType, metadata string
	err := r.db.QueryRowContext(ctx,
		`SELECT s.id, s.doc_id, s.ref, s.type, s.kind, s.title, s.content, s.content_hash, p.ref, s."order", s.content_type, s.metadata
		 FROM sections s
		 LEFT JOIN sections p ON s.parent_id = p.id
		 WHERE s.doc_id = ? AND s.ref = ?`, docID, ref,
	).Scan(&s.ID, &s.DocID, &refStr, &typStr, &kind, &s.Title, &s.Content, &s.ContentHash, &parentRefStr, &s.Order, &contentType, &metadata)
	if err == sql.ErrNoRows {
		return nil, core.ErrNotFound{Entity: "section", ID: ref}
	}
	if err != nil {
		return nil, fmt.Errorf("find section by ref: %w", err)
	}
	populateSection(&s, refStr, typStr, kind, parentRefStr, contentType, metadata)
	return &s, nil
}

func (r *DocumentRepo) FindSectionByRefGlobal(ctx context.Context, ref string) (*core.Section, string, error) {
	var s core.Section
	var refStr, typStr, kind string
	var parentRefStr sql.NullString
	var contentType, metadata string
	err := r.db.QueryRowContext(ctx,
		`SELECT s.id, s.doc_id, s.ref, s.type, s.kind, s.title, s.content, s.content_hash, p.ref, s."order", s.content_type, s.metadata
		 FROM sections s
		 LEFT JOIN sections p ON s.parent_id = p.id
		 WHERE s.ref = ? LIMIT 1`, ref,
	).Scan(&s.ID, &s.DocID, &refStr, &typStr, &kind, &s.Title, &s.Content, &s.ContentHash, &parentRefStr, &s.Order, &contentType, &metadata)
	if err == sql.ErrNoRows {
		return nil, "", core.ErrNotFound{Entity: "section", ID: ref}
	}
	if err != nil {
		return nil, "", fmt.Errorf("find section by ref %s: %w", ref, err)
	}
	populateSection(&s, refStr, typStr, kind, parentRefStr, contentType, metadata)
	return &s, s.DocID, nil
}

// FindSectionByID retrieves a section by its primary key.
func (r *DocumentRepo) FindSectionByID(ctx context.Context, sectionID string) (*core.Section, error) {
	var s core.Section
	var refStr, typStr, kind string
	var parentRefStr sql.NullString
	var contentType, metadata string
	err := r.db.QueryRowContext(ctx,
		`SELECT s.id, s.doc_id, s.ref, s.type, s.kind, s.title, s.content, s.content_hash, p.ref, s."order", s.content_type, s.metadata
		 FROM sections s
		 LEFT JOIN sections p ON s.parent_id = p.id
		 WHERE s.id = ?`, sectionID,
	).Scan(&s.ID, &s.DocID, &refStr, &typStr, &kind, &s.Title, &s.Content, &s.ContentHash, &parentRefStr, &s.Order, &contentType, &metadata)
	if err == sql.ErrNoRows {
		return nil, core.ErrNotFound{Entity: "section", ID: sectionID}
	}
	if err != nil {
		return nil, fmt.Errorf("find section by id %s: %w", sectionID, err)
	}
	populateSection(&s, refStr, typStr, kind, parentRefStr, contentType, metadata)
	return &s, nil
}

func (r *DocumentRepo) UpdateSectionContent(ctx context.Context, sectionID, content, contentHash string) error {
	return WithTx(ctx, r.db, func(tx *sql.Tx) error {
		var maxVer int
		err := tx.QueryRowContext(ctx,
			`SELECT COALESCE(MAX(version), 0) FROM section_versions WHERE section_id = ?`, sectionID,
		).Scan(&maxVer)
		if err != nil {
			return fmt.Errorf("query max version: %w", err)
		}
		nextVer := maxVer + 1

		versionID := core.NewID().String()
		_, err = tx.ExecContext(ctx,
			`INSERT INTO section_versions (id, section_id, version, content, content_hash, metadata)
			 VALUES (?, ?, ?, ?, ?, ?)`,
			versionID, sectionID, nextVer, content, contentHash, "{}",
		)
		if err != nil {
			return fmt.Errorf("insert section version: %w", err)
		}

		_, err = tx.ExecContext(ctx,
			`UPDATE sections SET content = ?, content_hash = ? WHERE id = ?`,
			content, contentHash, sectionID,
		)
		if err != nil {
			return fmt.Errorf("update section content: %w", err)
		}
		return nil
	})
}

func (r *DocumentRepo) GetSectionVersions(ctx context.Context, sectionID string) ([]core.SectionVersion, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, section_id, version, content, content_hash, metadata, created_at
		 FROM section_versions WHERE section_id = ? ORDER BY version`, sectionID,
	)
	if err != nil {
		return nil, fmt.Errorf("query section versions: %w", err)
	}
	defer rows.Close()

	var versions []core.SectionVersion
	for rows.Next() {
		var v core.SectionVersion
		var createdAt string
		if err := rows.Scan(&v.ID, &v.SectionID, &v.Version, &v.Content, &v.ContentHash, &v.Metadata, &createdAt); err != nil {
			return nil, fmt.Errorf("scan section version: %w", err)
		}
		v.CreatedAt = parseTime(createdAt)
		versions = append(versions, v)
	}
	return versions, rows.Err()
}

func (r *DocumentRepo) DeleteSection(ctx context.Context, sectionID string) error {
	return WithTx(ctx, r.db, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `DELETE FROM tags WHERE section_id = ?`, sectionID); err != nil {
			return fmt.Errorf("delete tags: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM section_versions WHERE section_id = ?`, sectionID); err != nil {
			return fmt.Errorf("delete section versions: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `UPDATE sections SET parent_id = NULL WHERE parent_id = ?`, sectionID); err != nil {
			return fmt.Errorf("nullify children: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM sections WHERE id = ?`, sectionID); err != nil {
			return fmt.Errorf("delete section: %w", err)
		}
		return nil
	})
}

func (r *DocumentRepo) AddTag(ctx context.Context, sectionID, tag string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO tags (section_id, tag) VALUES (?, ?) ON CONFLICT DO NOTHING`,
		sectionID, tag,
	)
	if err != nil {
		return fmt.Errorf("add tag: %w", err)
	}
	return nil
}

func (r *DocumentRepo) RemoveTag(ctx context.Context, sectionID, tag string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM tags WHERE section_id = ? AND tag = ?`,
		sectionID, tag,
	)
	if err != nil {
		return fmt.Errorf("remove tag: %w", err)
	}
	return nil
}

func (r *DocumentRepo) NextRefSeq(ctx context.Context, count int) (int, error) {
	var first int
	err := WithTx(ctx, r.db, func(tx *sql.Tx) error {
		var seq int
		if err := tx.QueryRowContext(ctx,
			`SELECT next_seq FROM ref_counter WHERE id = 1`,
		).Scan(&seq); err != nil {
			return fmt.Errorf("read ref_counter: %w", err)
		}
		first = seq
		_, err := tx.ExecContext(ctx,
			`UPDATE ref_counter SET next_seq = next_seq + ? WHERE id = 1`, count,
		)
		if err != nil {
			return fmt.Errorf("update ref_counter: %w", err)
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return first, nil
}

func (r *DocumentRepo) GetTags(ctx context.Context, sectionID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT tag FROM tags WHERE section_id = ? ORDER BY tag`, sectionID,
	)
	if err != nil {
		return nil, fmt.Errorf("query tags: %w", err)
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, fmt.Errorf("scan tag: %w", err)
		}
		tags = append(tags, tag)
	}
	return tags, rows.Err()
}

// CreateSections batch-inserts multiple sections in a single transaction.
func (r *DocumentRepo) CreateSections(ctx context.Context, sections []core.Section) error {
	return WithTx(ctx, r.db, func(tx *sql.Tx) error {
		stmt, err := tx.PrepareContext(ctx,
			`INSERT INTO sections (id, doc_id, ref, type, kind, title, content, content_hash, parent_id, "order", content_type, metadata)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
		if err != nil {
			return fmt.Errorf("prepare insert sections: %w", err)
		}
		defer stmt.Close()
		for i := range sections {
			s := &sections[i]
			var parentID *string
			if s.ParentRef != nil {
				var pid string
				err := tx.QueryRowContext(ctx,
					`SELECT id FROM sections WHERE doc_id = ? AND ref = ?`,
					s.DocID, s.ParentRef.String(),
				).Scan(&pid)
				if err == nil {
					parentID = &pid
				}
			}
			contentType := string(s.ContentType)
			if contentType == "" {
				contentType = string(core.ContentNative)
			}
			metadata := s.Metadata
			if metadata == "" {
				metadata = "{}"
			}
			if _, err := stmt.ExecContext(ctx, s.ID, s.DocID, s.Ref.String(), string(s.Type), s.Kind, s.Title, s.Content, s.ContentHash, parentID, s.Order, contentType, metadata); err != nil {
				return fmt.Errorf("insert section %s: %w", s.ID, err)
			}
		}
		return nil
	})
}

// ListDocumentsWithSectionCounts returns all documents with their section counts.
func (r *DocumentRepo) ListDocumentsWithSectionCounts(ctx context.Context) ([]core.DocumentSummary, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT d.id, d.title, d.owner_id, d.status, d.source, d.doc_type, d.parent_doc_id, d.created_at, COUNT(s.id)
		 FROM documents d LEFT JOIN sections s ON d.id = s.doc_id
		 GROUP BY d.id ORDER BY d.created_at`)
	if err != nil {
		return nil, fmt.Errorf("list documents with counts: %w", err)
	}
	defer rows.Close()

	var summaries []core.DocumentSummary
	for rows.Next() {
		var doc core.Document
		var createdAt string
		var parentDocID sql.NullString
		var count int
		if err := rows.Scan(&doc.ID, &doc.Title, &doc.OwnerID, &doc.Status, &doc.Source, &doc.DocType, &parentDocID, &createdAt, &count); err != nil {
			return nil, fmt.Errorf("scan document summary: %w", err)
		}
		if parentDocID.Valid {
			doc.ParentDocID = parentDocID.String
		}
		doc.CreatedAt = parseTime(createdAt)
		summaries = append(summaries, core.DocumentSummary{Document: &doc, SectionCount: count})
	}
	return summaries, rows.Err()
}

// DeleteDocument removes a document and all its sections, section_versions, and tags.
func (r *DocumentRepo) DeleteDocument(ctx context.Context, docID string) error {
	return WithTx(ctx, r.db, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx,
			`DELETE FROM tags WHERE section_id IN (SELECT id FROM sections WHERE doc_id = ?)`, docID); err != nil {
			return fmt.Errorf("delete tags: %w", err)
		}
		if _, err := tx.ExecContext(ctx,
			`DELETE FROM section_versions WHERE section_id IN (SELECT id FROM sections WHERE doc_id = ?)`, docID); err != nil {
			return fmt.Errorf("delete section versions: %w", err)
		}
		if _, err := tx.ExecContext(ctx,
			`DELETE FROM sections WHERE doc_id = ?`, docID); err != nil {
			return fmt.Errorf("delete sections: %w", err)
		}
		res, err := tx.ExecContext(ctx,
			`DELETE FROM documents WHERE id = ?`, docID)
		if err != nil {
			return fmt.Errorf("delete document: %w", err)
		}
		n, err := res.RowsAffected()
		if err != nil {
			return fmt.Errorf("rows affected: %w", err)
		}
		if n == 0 {
			return core.ErrNotFound{Entity: "document", ID: docID}
		}
		return nil
	})
}

// ArchiveDocument sets a document's status to archived.
func (r *DocumentRepo) ArchiveDocument(ctx context.Context, docID string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE documents SET status = ? WHERE id = ?`, string(core.DocumentArchived), docID)
	if err != nil {
		return fmt.Errorf("archive document: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if n == 0 {
		return core.ErrNotFound{Entity: "document", ID: docID}
	}
	return nil
}

// FindSectionsByTag returns all sections tagged with the given tag string.
func (r *DocumentRepo) FindSectionsByTag(ctx context.Context, tag string) ([]*core.Section, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT s.id, s.doc_id, s.ref, s.type, s.kind, s.title, s.content, s.content_hash, p.ref, s."order", s.content_type, s.metadata
		 FROM sections s
		 LEFT JOIN sections p ON s.parent_id = p.id
		 INNER JOIN tags t ON s.id = t.section_id
		 WHERE t.tag = ?
		 ORDER BY s."order"`, tag)
	if err != nil {
		return nil, fmt.Errorf("find sections by tag: %w", err)
	}
	defer rows.Close()

	var sections []*core.Section
	for rows.Next() {
		var s core.Section
		var refStr, typStr, kind string
		var parentRefStr sql.NullString
		var contentType, metadata string
		if err := rows.Scan(&s.ID, &s.DocID, &refStr, &typStr, &kind, &s.Title, &s.Content, &s.ContentHash, &parentRefStr, &s.Order, &contentType, &metadata); err != nil {
			return nil, fmt.Errorf("scan section by tag: %w", err)
		}
		populateSection(&s, refStr, typStr, kind, parentRefStr, contentType, metadata)
		sections = append(sections, &s)
	}
	return sections, rows.Err()
}

// SearchSections performs full-text search on sections using FTS5.
func (r *DocumentRepo) SearchSections(ctx context.Context, query string) ([]*core.Section, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT s.id, s.doc_id, s.ref, s.type, s.kind, s.title, s.content, s.content_hash, p.ref, s."order", s.content_type, s.metadata
		 FROM sections s
		 LEFT JOIN sections p ON s.parent_id = p.id
		 INNER JOIN sections_fts ON sections_fts.rowid = s.rowid
		 WHERE sections_fts MATCH ?
		 ORDER BY rank`, query)
	if err != nil {
		return nil, fmt.Errorf("search sections: %w", err)
	}
	defer rows.Close()

	var sections []*core.Section
	for rows.Next() {
		var s core.Section
		var refStr, typStr, kind string
		var parentRefStr sql.NullString
		var contentType, metadata string
		if err := rows.Scan(&s.ID, &s.DocID, &refStr, &typStr, &kind, &s.Title, &s.Content, &s.ContentHash, &parentRefStr, &s.Order, &contentType, &metadata); err != nil {
			return nil, fmt.Errorf("scan search section: %w", err)
		}
		populateSection(&s, refStr, typStr, kind, parentRefStr, contentType, metadata)
		sections = append(sections, &s)
	}
	return sections, rows.Err()
}
