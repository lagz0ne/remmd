package store_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/lagz0ne/remmd/internal/core"
	"github.com/lagz0ne/remmd/internal/store"
)

func setupSubTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := store.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("open in-memory db: %v", err)
	}
	if err := store.Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	t.Cleanup(func() { store.CloseDB(db) })
	return db
}

func TestCreateSubscription_And_FindByID(t *testing.T) {
	t.Parallel()
	db := setupSubTestDB(t)
	repo := store.NewSubscriptionRepo(db)
	ctx := context.Background()

	sub := core.NewSubscription("@a1", "api-spec", "user-1")
	if err := repo.CreateSubscription(ctx, sub); err != nil {
		t.Fatalf("CreateSubscription: %v", err)
	}

	got, err := repo.FindByID(ctx, sub.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.ID != sub.ID {
		t.Errorf("ID = %q, want %q", got.ID, sub.ID)
	}
	if got.SubscriberRef != "@a1" {
		t.Errorf("SubscriberRef = %q, want %q", got.SubscriberRef, "@a1")
	}
	if got.TagExpression != "api-spec" {
		t.Errorf("TagExpression = %q, want %q", got.TagExpression, "api-spec")
	}
	if got.Status != core.SubActive {
		t.Errorf("Status = %q, want %q", got.Status, core.SubActive)
	}
	if got.CreatedBy != "user-1" {
		t.Errorf("CreatedBy = %q, want %q", got.CreatedBy, "user-1")
	}
	if got.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
}

func TestFindByID_NotFound(t *testing.T) {
	t.Parallel()
	db := setupSubTestDB(t)
	repo := store.NewSubscriptionRepo(db)
	ctx := context.Background()

	_, err := repo.FindByID(ctx, "does-not-exist")
	if err == nil {
		t.Fatal("expected error for non-existent subscription")
	}
	if !errors.Is(err, core.ErrNotFound{}) {
		t.Errorf("expected ErrNotFound, got %T: %v", err, err)
	}
}

func TestListActiveSubscriptions(t *testing.T) {
	t.Parallel()
	db := setupSubTestDB(t)
	repo := store.NewSubscriptionRepo(db)
	ctx := context.Background()

	// Create 2 active subscriptions
	sub1 := core.NewSubscription("@a1", "tag-a", "user-1")
	sub2 := core.NewSubscription("@b1", "tag-b", "user-1")
	if err := repo.CreateSubscription(ctx, sub1); err != nil {
		t.Fatalf("CreateSubscription sub1: %v", err)
	}
	if err := repo.CreateSubscription(ctx, sub2); err != nil {
		t.Fatalf("CreateSubscription sub2: %v", err)
	}

	// Archive one
	if err := repo.ArchiveSubscription(ctx, sub1.ID); err != nil {
		t.Fatalf("ArchiveSubscription: %v", err)
	}

	// Only sub2 should be active
	active, err := repo.ListActiveSubscriptions(ctx)
	if err != nil {
		t.Fatalf("ListActiveSubscriptions: %v", err)
	}
	if len(active) != 1 {
		t.Fatalf("expected 1 active subscription, got %d", len(active))
	}
	if active[0].ID != sub2.ID {
		t.Errorf("active subscription ID = %q, want %q", active[0].ID, sub2.ID)
	}
}

func TestFindByTag(t *testing.T) {
	t.Parallel()
	db := setupSubTestDB(t)
	repo := store.NewSubscriptionRepo(db)
	ctx := context.Background()

	sub1 := core.NewSubscription("@a1", "api-spec", "user-1")
	sub2 := core.NewSubscription("@b1", "api-spec", "user-2")
	sub3 := core.NewSubscription("@c1", "design-doc", "user-1")
	for _, s := range []*core.Subscription{sub1, sub2, sub3} {
		if err := repo.CreateSubscription(ctx, s); err != nil {
			t.Fatalf("CreateSubscription: %v", err)
		}
	}

	found, err := repo.FindByTag(ctx, "api-spec")
	if err != nil {
		t.Fatalf("FindByTag: %v", err)
	}
	if len(found) != 2 {
		t.Fatalf("expected 2 subscriptions for 'api-spec', got %d", len(found))
	}

	found2, err := repo.FindByTag(ctx, "design-doc")
	if err != nil {
		t.Fatalf("FindByTag: %v", err)
	}
	if len(found2) != 1 {
		t.Fatalf("expected 1 subscription for 'design-doc', got %d", len(found2))
	}
}

func TestArchiveSubscription(t *testing.T) {
	t.Parallel()
	db := setupSubTestDB(t)
	repo := store.NewSubscriptionRepo(db)
	ctx := context.Background()

	sub := core.NewSubscription("@a1", "api-spec", "user-1")
	if err := repo.CreateSubscription(ctx, sub); err != nil {
		t.Fatalf("CreateSubscription: %v", err)
	}

	if err := repo.ArchiveSubscription(ctx, sub.ID); err != nil {
		t.Fatalf("ArchiveSubscription: %v", err)
	}

	got, err := repo.FindByID(ctx, sub.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.Status != core.SubArchived {
		t.Errorf("Status = %q, want %q", got.Status, core.SubArchived)
	}
}

func TestCreateFire_And_ListPendingFires(t *testing.T) {
	t.Parallel()
	db := setupSubTestDB(t)
	repo := store.NewSubscriptionRepo(db)
	ctx := context.Background()

	sub := core.NewSubscription("@a1", "api-spec", "user-1")
	if err := repo.CreateSubscription(ctx, sub); err != nil {
		t.Fatalf("CreateSubscription: %v", err)
	}

	fire := &core.SubscriptionFire{
		ID:              core.NewID().String(),
		SubscriptionID:  sub.ID,
		TriggeringDocID: "doc-1",
		TriggeringRef:   "@c1",
		MatchedTag:      "api-spec",
		Status:          core.FirePending,
		FiredAt:         time.Now(),
	}
	if err := repo.CreateFire(ctx, fire); err != nil {
		t.Fatalf("CreateFire: %v", err)
	}

	fires, err := repo.ListPendingFires(ctx, sub.ID)
	if err != nil {
		t.Fatalf("ListPendingFires: %v", err)
	}
	if len(fires) != 1 {
		t.Fatalf("expected 1 pending fire, got %d", len(fires))
	}
	if fires[0].ID != fire.ID {
		t.Errorf("fire ID = %q, want %q", fires[0].ID, fire.ID)
	}
	if fires[0].TriggeringDocID != "doc-1" {
		t.Errorf("TriggeringDocID = %q, want %q", fires[0].TriggeringDocID, "doc-1")
	}
	if fires[0].TriggeringRef != "@c1" {
		t.Errorf("TriggeringRef = %q, want %q", fires[0].TriggeringRef, "@c1")
	}
	if fires[0].MatchedTag != "api-spec" {
		t.Errorf("MatchedTag = %q, want %q", fires[0].MatchedTag, "api-spec")
	}
}

func TestDismissFire(t *testing.T) {
	t.Parallel()
	db := setupSubTestDB(t)
	repo := store.NewSubscriptionRepo(db)
	ctx := context.Background()

	sub := core.NewSubscription("@a1", "api-spec", "user-1")
	if err := repo.CreateSubscription(ctx, sub); err != nil {
		t.Fatalf("CreateSubscription: %v", err)
	}

	fire := &core.SubscriptionFire{
		ID:              core.NewID().String(),
		SubscriptionID:  sub.ID,
		TriggeringDocID: "doc-1",
		TriggeringRef:   "@c1",
		MatchedTag:      "api-spec",
		Status:          core.FirePending,
		FiredAt:         time.Now(),
	}
	if err := repo.CreateFire(ctx, fire); err != nil {
		t.Fatalf("CreateFire: %v", err)
	}

	if err := repo.DismissFire(ctx, fire.ID); err != nil {
		t.Fatalf("DismissFire: %v", err)
	}

	// Verify the fire is now dismissed by re-querying
	fires, err := repo.ListPendingFires(ctx, sub.ID)
	if err != nil {
		t.Fatalf("ListPendingFires: %v", err)
	}
	if len(fires) != 0 {
		t.Errorf("expected 0 pending fires after dismiss, got %d", len(fires))
	}
}

func TestListPendingFires_ExcludesDismissed(t *testing.T) {
	t.Parallel()
	db := setupSubTestDB(t)
	repo := store.NewSubscriptionRepo(db)
	ctx := context.Background()

	sub := core.NewSubscription("@a1", "api-spec", "user-1")
	if err := repo.CreateSubscription(ctx, sub); err != nil {
		t.Fatalf("CreateSubscription: %v", err)
	}

	// Create two fires
	fire1 := &core.SubscriptionFire{
		ID:              core.NewID().String(),
		SubscriptionID:  sub.ID,
		TriggeringDocID: "doc-1",
		TriggeringRef:   "@c1",
		MatchedTag:      "api-spec",
		Status:          core.FirePending,
		FiredAt:         time.Now(),
	}
	fire2 := &core.SubscriptionFire{
		ID:              core.NewID().String(),
		SubscriptionID:  sub.ID,
		TriggeringDocID: "doc-2",
		TriggeringRef:   "@d1",
		MatchedTag:      "api-spec",
		Status:          core.FirePending,
		FiredAt:         time.Now(),
	}
	if err := repo.CreateFire(ctx, fire1); err != nil {
		t.Fatalf("CreateFire fire1: %v", err)
	}
	if err := repo.CreateFire(ctx, fire2); err != nil {
		t.Fatalf("CreateFire fire2: %v", err)
	}

	// Dismiss fire1
	if err := repo.DismissFire(ctx, fire1.ID); err != nil {
		t.Fatalf("DismissFire: %v", err)
	}

	// Only fire2 should remain pending
	fires, err := repo.ListPendingFires(ctx, sub.ID)
	if err != nil {
		t.Fatalf("ListPendingFires: %v", err)
	}
	if len(fires) != 1 {
		t.Fatalf("expected 1 pending fire, got %d", len(fires))
	}
	if fires[0].ID != fire2.ID {
		t.Errorf("remaining fire ID = %q, want %q", fires[0].ID, fire2.ID)
	}
}
