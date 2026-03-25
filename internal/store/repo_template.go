package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lagz0ne/remmd/internal/core"
)

// TemplateRepo is the SQLite implementation of core.TemplateRepository.
type TemplateRepo struct {
	db *sql.DB
}

// NewTemplateRepo creates a TemplateRepo backed by the given *sql.DB.
func NewTemplateRepo(db *sql.DB) *TemplateRepo {
	return &TemplateRepo{db: db}
}

// SetTemplate upserts a schema template (required section kind for a document type).
func (r *TemplateRepo) SetTemplate(ctx context.Context, t core.SchemaTemplate) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO schema_templates (doc_type, required_kind, min_count)
		 VALUES (?, ?, ?)
		 ON CONFLICT(doc_type, required_kind) DO UPDATE SET min_count = excluded.min_count`,
		t.DocType, t.RequiredKind, t.MinCount,
	)
	if err != nil {
		return fmt.Errorf("set template: %w", err)
	}
	return nil
}

// GetTemplates returns all schema templates for a given document type.
func (r *TemplateRepo) GetTemplates(ctx context.Context, docType string) ([]core.SchemaTemplate, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT doc_type, required_kind, min_count
		 FROM schema_templates WHERE doc_type = ? ORDER BY required_kind`, docType,
	)
	if err != nil {
		return nil, fmt.Errorf("get templates: %w", err)
	}
	defer rows.Close()

	var templates []core.SchemaTemplate
	for rows.Next() {
		var t core.SchemaTemplate
		if err := rows.Scan(&t.DocType, &t.RequiredKind, &t.MinCount); err != nil {
			return nil, fmt.Errorf("scan template: %w", err)
		}
		templates = append(templates, t)
	}
	return templates, rows.Err()
}

// DeleteTemplate removes a schema template by document type and required kind.
func (r *TemplateRepo) DeleteTemplate(ctx context.Context, docType, requiredKind string) error {
	res, err := r.db.ExecContext(ctx,
		`DELETE FROM schema_templates WHERE doc_type = ? AND required_kind = ?`,
		docType, requiredKind,
	)
	if err != nil {
		return fmt.Errorf("delete template: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return core.ErrNotFound{Entity: "template", ID: docType + "/" + requiredKind}
	}
	return nil
}
