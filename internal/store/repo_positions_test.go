package store_test

import (
	"context"
	"testing"

	"github.com/lagz0ne/remmd/internal/store"
)

func TestPositionStore_SaveAndLoad(t *testing.T) {
	t.Parallel()
	db := setupDocTestDB(t)
	ps := store.NewPositionStore(db)
	ctx := context.Background()

	// Empty initially
	positions, err := ps.LoadPositions(ctx)
	if err != nil {
		t.Fatalf("LoadPositions: %v", err)
	}
	if len(positions) != 0 {
		t.Errorf("expected 0 positions, got %d", len(positions))
	}

	// Save some positions
	err = ps.SavePositions(ctx, []store.NodePosition{
		{NodeID: "doc-1", X: 100.5, Y: 200.3},
		{NodeID: "doc-2", X: -50.0, Y: 300.7},
	})
	if err != nil {
		t.Fatalf("SavePositions: %v", err)
	}

	positions, err = ps.LoadPositions(ctx)
	if err != nil {
		t.Fatalf("LoadPositions after save: %v", err)
	}
	if len(positions) != 2 {
		t.Fatalf("expected 2 positions, got %d", len(positions))
	}

	p1 := positions["doc-1"]
	if p1.X != 100.5 || p1.Y != 200.3 {
		t.Errorf("doc-1 position = (%f, %f), want (100.5, 200.3)", p1.X, p1.Y)
	}
	p2 := positions["doc-2"]
	if p2.X != -50.0 || p2.Y != 300.7 {
		t.Errorf("doc-2 position = (%f, %f), want (-50.0, 300.7)", p2.X, p2.Y)
	}
}

func TestPositionStore_Upsert(t *testing.T) {
	t.Parallel()
	db := setupDocTestDB(t)
	ps := store.NewPositionStore(db)
	ctx := context.Background()

	// Save initial position
	err := ps.SavePositions(ctx, []store.NodePosition{
		{NodeID: "doc-1", X: 10, Y: 20},
	})
	if err != nil {
		t.Fatalf("SavePositions: %v", err)
	}

	// Update position
	err = ps.SavePositions(ctx, []store.NodePosition{
		{NodeID: "doc-1", X: 99, Y: 88},
	})
	if err != nil {
		t.Fatalf("SavePositions update: %v", err)
	}

	positions, err := ps.LoadPositions(ctx)
	if err != nil {
		t.Fatalf("LoadPositions: %v", err)
	}
	if len(positions) != 1 {
		t.Fatalf("expected 1 position, got %d", len(positions))
	}
	p := positions["doc-1"]
	if p.X != 99 || p.Y != 88 {
		t.Errorf("position = (%f, %f), want (99, 88)", p.X, p.Y)
	}
}

func TestPositionStore_Clear(t *testing.T) {
	t.Parallel()
	db := setupDocTestDB(t)
	ps := store.NewPositionStore(db)
	ctx := context.Background()

	err := ps.SavePositions(ctx, []store.NodePosition{
		{NodeID: "doc-1", X: 10, Y: 20},
		{NodeID: "doc-2", X: 30, Y: 40},
	})
	if err != nil {
		t.Fatalf("SavePositions: %v", err)
	}

	err = ps.ClearPositions(ctx)
	if err != nil {
		t.Fatalf("ClearPositions: %v", err)
	}

	positions, err := ps.LoadPositions(ctx)
	if err != nil {
		t.Fatalf("LoadPositions: %v", err)
	}
	if len(positions) != 0 {
		t.Errorf("expected 0 positions after clear, got %d", len(positions))
	}
}
