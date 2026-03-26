package store

import (
	"context"
	"database/sql"
	"fmt"
)

// NodePosition represents a persisted canvas position for a node.
type NodePosition struct {
	NodeID string  `json:"node_id"`
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
}

// PositionStore handles node position persistence.
type PositionStore struct {
	db *sql.DB
}

// NewPositionStore creates a PositionStore backed by the given *sql.DB.
func NewPositionStore(db *sql.DB) *PositionStore {
	return &PositionStore{db: db}
}

// SavePositions upserts a batch of node positions.
func (s *PositionStore) SavePositions(ctx context.Context, positions []NodePosition) error {
	return WithTx(ctx, s.db, func(tx *sql.Tx) error {
		stmt, err := tx.PrepareContext(ctx,
			`INSERT INTO node_positions (node_id, x, y, updated_at)
			 VALUES (?, ?, ?, datetime('now'))
			 ON CONFLICT(node_id) DO UPDATE SET x=excluded.x, y=excluded.y, updated_at=excluded.updated_at`)
		if err != nil {
			return fmt.Errorf("prepare upsert: %w", err)
		}
		defer stmt.Close()

		for _, p := range positions {
			if _, err := stmt.ExecContext(ctx, p.NodeID, p.X, p.Y); err != nil {
				return fmt.Errorf("upsert position %q: %w", p.NodeID, err)
			}
		}
		return nil
	})
}

// LoadPositions returns all saved positions keyed by node_id.
func (s *PositionStore) LoadPositions(ctx context.Context) (map[string]NodePosition, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT node_id, x, y FROM node_positions`)
	if err != nil {
		return nil, fmt.Errorf("query positions: %w", err)
	}
	defer rows.Close()

	result := make(map[string]NodePosition)
	for rows.Next() {
		var p NodePosition
		if err := rows.Scan(&p.NodeID, &p.X, &p.Y); err != nil {
			return nil, fmt.Errorf("scan position: %w", err)
		}
		result[p.NodeID] = p
	}
	return result, rows.Err()
}

// ClearPositions deletes all saved positions.
func (s *PositionStore) ClearPositions(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM node_positions`)
	return err
}
