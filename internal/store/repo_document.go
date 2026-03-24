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
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO documents (id, title, owner_id, status, source, created_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		doc.ID, doc.Title, doc.OwnerID, doc.Status, doc.Source,
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
	err := r.db.QueryRowContext(ctx,
		`SELECT id, title, owner_id, status, source, created_at FROM documents WHERE id = ?`, id,
	).Scan(&doc.ID, &doc.Title, &doc.OwnerID, &doc.Status, &doc.Source, &createdAt)
	if err == sql.ErrNoRows {
		return nil, core.ErrNotFound{Entity: "document", ID: id}
	}
	if err != nil {
		return nil, fmt.Errorf("find document %s: %w", id, err)
	}
	doc.CreatedAt = parseTime(createdAt)
	return &doc, nil
}

func (r *DocumentRepo) ListDocuments(ctx context.Context) ([]*core.Document, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, title, owner_id, status, source, created_at FROM documents ORDER BY created_at`)
	if err != nil {
		return nil, fmt.Errorf("list documents: %w", err)
	}
	defer rows.Close()

	var docs []*core.Document
	for rows.Next() {
		var doc core.Document
		var createdAt string
		if err := rows.Scan(&doc.ID, &doc.Title, &doc.OwnerID, &doc.Status, &doc.Source, &createdAt); err != nil {
			return nil, fmt.Errorf("scan document: %w", err)
		}
		doc.CreatedAt = parseTime(createdAt)
		docs = append(docs, &doc)
	}
	return docs, rows.Err()
}

// populateSection sets parsed fields on a Section from raw scanned strings.
func populateSection(s *core.Section, refStr, typStr string, parentRefStr sql.NullString, contentType, metadata string) {
	ref, _ := core.ParseRef(refStr)
	s.Ref = ref
	s.Type = core.SectionType(typStr)
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
		`INSERT INTO sections (id, doc_id, ref, type, title, content, content_hash, parent_id, "order", content_type, metadata)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		s.ID, s.DocID, s.Ref.String(), string(s.Type), s.Title, s.Content, s.ContentHash, parentID, s.Order, contentType, metadata,
	)
	if err != nil {
		return fmt.Errorf("insert section: %w", err)
	}
	return nil
}

func (r *DocumentRepo) ListSections(ctx context.Context, docID string) ([]*core.Section, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT s.id, s.doc_id, s.ref, s.type, s.title, s.content, s.content_hash, p.ref, s."order", s.content_type, s.metadata
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
		var refStr string
		var typStr string
		var parentRefStr sql.NullString
		var contentType, metadata string
		if err := rows.Scan(&s.ID, &s.DocID, &refStr, &typStr, &s.Title, &s.Content, &s.ContentHash, &parentRefStr, &s.Order, &contentType, &metadata); err != nil {
			return nil, fmt.Errorf("scan section: %w", err)
		}
		populateSection(&s, refStr, typStr, parentRefStr, contentType, metadata)
		sections = append(sections, &s)
	}
	return sections, rows.Err()
}

func (r *DocumentRepo) FindSectionByRef(ctx context.Context, docID string, ref string) (*core.Section, error) {
	var s core.Section
	var refStr, typStr string
	var parentRefStr sql.NullString
	var contentType, metadata string
	err := r.db.QueryRowContext(ctx,
		`SELECT s.id, s.doc_id, s.ref, s.type, s.title, s.content, s.content_hash, p.ref, s."order", s.content_type, s.metadata
		 FROM sections s
		 LEFT JOIN sections p ON s.parent_id = p.id
		 WHERE s.doc_id = ? AND s.ref = ?`, docID, ref,
	).Scan(&s.ID, &s.DocID, &refStr, &typStr, &s.Title, &s.Content, &s.ContentHash, &parentRefStr, &s.Order, &contentType, &metadata)
	if err == sql.ErrNoRows {
		return nil, core.ErrNotFound{Entity: "section", ID: ref}
	}
	if err != nil {
		return nil, fmt.Errorf("find section by ref: %w", err)
	}
	populateSection(&s, refStr, typStr, parentRefStr, contentType, metadata)
	return &s, nil
}

func (r *DocumentRepo) FindSectionByRefGlobal(ctx context.Context, ref string) (*core.Section, string, error) {
	var s core.Section
	var refStr, typStr string
	var parentRefStr sql.NullString
	var contentType, metadata string
	err := r.db.QueryRowContext(ctx,
		`SELECT s.id, s.doc_id, s.ref, s.type, s.title, s.content, s.content_hash, p.ref, s."order", s.content_type, s.metadata
		 FROM sections s
		 LEFT JOIN sections p ON s.parent_id = p.id
		 WHERE s.ref = ? LIMIT 1`, ref,
	).Scan(&s.ID, &s.DocID, &refStr, &typStr, &s.Title, &s.Content, &s.ContentHash, &parentRefStr, &s.Order, &contentType, &metadata)
	if err == sql.ErrNoRows {
		return nil, "", core.ErrNotFound{Entity: "section", ID: ref}
	}
	if err != nil {
		return nil, "", fmt.Errorf("find section by ref %s: %w", ref, err)
	}
	populateSection(&s, refStr, typStr, parentRefStr, contentType, metadata)
	return &s, s.DocID, nil
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
			`INSERT INTO section_versions (id, section_id, version, content, content_hash)
			 VALUES (?, ?, ?, ?, ?)`,
			versionID, sectionID, nextVer, content, contentHash,
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
		`SELECT id, section_id, version, content, content_hash, created_at
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
		if err := rows.Scan(&v.ID, &v.SectionID, &v.Version, &v.Content, &v.ContentHash, &createdAt); err != nil {
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
