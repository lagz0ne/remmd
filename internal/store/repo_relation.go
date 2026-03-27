package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lagz0ne/remmd/internal/core"
)

// RelationRepo is the SQLite implementation of core.RelationRepository.
type RelationRepo struct {
	db *sql.DB
}

// NewRelationRepo creates a RelationRepo backed by the given *sql.DB.
func NewRelationRepo(db *sql.DB) *RelationRepo {
	return &RelationRepo{db: db}
}

// CreateRelation persists a new relation.
func (r *RelationRepo) CreateRelation(ctx context.Context, rel *core.Relation) error {
	if rel.CreatedAt.IsZero() {
		rel.CreatedAt = time.Now()
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO relations (id, from_doc_id, to_doc_id, relation_type, metadata, created_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		rel.ID, rel.FromDocID, rel.ToDocID, rel.RelationType, rel.Metadata,
		formatTime(rel.CreatedAt),
	)
	if err != nil {
		return fmt.Errorf("insert relation: %w", err)
	}
	return nil
}

// ListRelationsFrom returns all relations originating from the given document.
func (r *RelationRepo) ListRelationsFrom(ctx context.Context, docID string) ([]core.Relation, error) {
	return r.listRelations(ctx,
		`SELECT id, from_doc_id, to_doc_id, relation_type, metadata, created_at
		 FROM relations WHERE from_doc_id = ? ORDER BY created_at`, docID)
}

// ListRelationsTo returns all relations pointing to the given document.
func (r *RelationRepo) ListRelationsTo(ctx context.Context, docID string) ([]core.Relation, error) {
	return r.listRelations(ctx,
		`SELECT id, from_doc_id, to_doc_id, relation_type, metadata, created_at
		 FROM relations WHERE to_doc_id = ? ORDER BY created_at`, docID)
}

// ListAllRelations returns every relation in the store.
func (r *RelationRepo) ListAllRelations(ctx context.Context) ([]core.Relation, error) {
	return r.listRelations(ctx,
		`SELECT id, from_doc_id, to_doc_id, relation_type, metadata, created_at
		 FROM relations ORDER BY created_at`)
}

// DeleteRelation removes a relation by ID.
func (r *RelationRepo) DeleteRelation(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM relations WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete relation: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return core.ErrNotFound{Entity: "relation", ID: id}
	}
	return nil
}

func (r *RelationRepo) listRelations(ctx context.Context, query string, args ...any) ([]core.Relation, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list relations: %w", err)
	}
	defer rows.Close()

	var relations []core.Relation
	for rows.Next() {
		var rel core.Relation
		var createdAt string
		if err := rows.Scan(&rel.ID, &rel.FromDocID, &rel.ToDocID, &rel.RelationType, &rel.Metadata, &createdAt); err != nil {
			return nil, fmt.Errorf("scan relation: %w", err)
		}
		rel.CreatedAt = parseTime(createdAt)
		relations = append(relations, rel)
	}
	return relations, rows.Err()
}
