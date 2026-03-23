package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lagz0ne/remmd/internal/core"
)

// SubscriptionRepo is the SQLite implementation of subscription persistence.
type SubscriptionRepo struct {
	db *sql.DB
}

// NewSubscriptionRepo creates a SubscriptionRepo backed by the given *sql.DB.
func NewSubscriptionRepo(db *sql.DB) *SubscriptionRepo {
	return &SubscriptionRepo{db: db}
}

// CreateSubscription persists a new subscription.
func (r *SubscriptionRepo) CreateSubscription(ctx context.Context, sub *core.Subscription) error {
	if sub.CreatedAt.IsZero() {
		sub.CreatedAt = time.Now()
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO subscriptions (id, subscriber_ref, tag_expression, status, created_by, created_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		sub.ID, sub.SubscriberRef, sub.TagExpression, sub.Status, sub.CreatedBy,
		formatTime(sub.CreatedAt),
	)
	if err != nil {
		return fmt.Errorf("insert subscription: %w", err)
	}
	return nil
}

// FindByID retrieves a subscription by ID.
func (r *SubscriptionRepo) FindByID(ctx context.Context, id string) (*core.Subscription, error) {
	var sub core.Subscription
	var createdAt string
	err := r.db.QueryRowContext(ctx,
		`SELECT id, subscriber_ref, tag_expression, status, created_by, created_at
		 FROM subscriptions WHERE id = ?`, id,
	).Scan(&sub.ID, &sub.SubscriberRef, &sub.TagExpression, &sub.Status, &sub.CreatedBy, &createdAt)
	if err == sql.ErrNoRows {
		return nil, core.ErrNotFound{Entity: "subscription", ID: id}
	}
	if err != nil {
		return nil, fmt.Errorf("find subscription %s: %w", id, err)
	}
	sub.CreatedAt = parseTime(createdAt)
	return &sub, nil
}

// ListActiveSubscriptions returns all subscriptions with status "active".
func (r *SubscriptionRepo) ListActiveSubscriptions(ctx context.Context) ([]*core.Subscription, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, subscriber_ref, tag_expression, status, created_by, created_at
		 FROM subscriptions WHERE status = 'active' ORDER BY created_at`)
	if err != nil {
		return nil, fmt.Errorf("list active subscriptions: %w", err)
	}
	defer rows.Close()

	var subs []*core.Subscription
	for rows.Next() {
		sub, err := scanSubscription(rows)
		if err != nil {
			return nil, err
		}
		subs = append(subs, sub)
	}
	return subs, rows.Err()
}

// FindByTag returns all active subscriptions matching the given tag.
func (r *SubscriptionRepo) FindByTag(ctx context.Context, tag string) ([]*core.Subscription, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, subscriber_ref, tag_expression, status, created_by, created_at
		 FROM subscriptions WHERE tag_expression = ? AND status = 'active' ORDER BY created_at`, tag)
	if err != nil {
		return nil, fmt.Errorf("find subscriptions by tag: %w", err)
	}
	defer rows.Close()

	var subs []*core.Subscription
	for rows.Next() {
		sub, err := scanSubscription(rows)
		if err != nil {
			return nil, err
		}
		subs = append(subs, sub)
	}
	return subs, rows.Err()
}

// ArchiveSubscription sets the subscription status to "archived".
func (r *SubscriptionRepo) ArchiveSubscription(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE subscriptions SET status = 'archived' WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("archive subscription: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return core.ErrNotFound{Entity: "subscription", ID: id}
	}
	return nil
}

// CreateFire persists a new subscription fire.
func (r *SubscriptionRepo) CreateFire(ctx context.Context, fire *core.SubscriptionFire) error {
	if fire.FiredAt.IsZero() {
		fire.FiredAt = time.Now()
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO subscription_fires (id, subscription_id, triggering_doc_id, triggering_ref, matched_tag, status, fired_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		fire.ID, fire.SubscriptionID, fire.TriggeringDocID, fire.TriggeringRef, fire.MatchedTag, fire.Status,
		formatTime(fire.FiredAt),
	)
	if err != nil {
		return fmt.Errorf("insert subscription fire: %w", err)
	}
	return nil
}

// ListPendingFires returns all fires with status "pending" for a subscription.
func (r *SubscriptionRepo) ListPendingFires(ctx context.Context, subscriptionID string) ([]*core.SubscriptionFire, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, subscription_id, triggering_doc_id, triggering_ref, matched_tag, status, fired_at
		 FROM subscription_fires WHERE subscription_id = ? AND status = 'pending' ORDER BY fired_at`, subscriptionID)
	if err != nil {
		return nil, fmt.Errorf("list pending fires: %w", err)
	}
	defer rows.Close()

	var fires []*core.SubscriptionFire
	for rows.Next() {
		fire, err := scanFire(rows)
		if err != nil {
			return nil, err
		}
		fires = append(fires, fire)
	}
	return fires, rows.Err()
}

// DismissFire sets a fire's status to "dismissed".
func (r *SubscriptionRepo) DismissFire(ctx context.Context, fireID string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE subscription_fires SET status = 'dismissed' WHERE id = ?`, fireID)
	if err != nil {
		return fmt.Errorf("dismiss fire: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return core.ErrNotFound{Entity: "subscription_fire", ID: fireID}
	}
	return nil
}

func scanSubscription(row scannable) (*core.Subscription, error) {
	var sub core.Subscription
	var createdAt string
	if err := row.Scan(&sub.ID, &sub.SubscriberRef, &sub.TagExpression, &sub.Status, &sub.CreatedBy, &createdAt); err != nil {
		return nil, fmt.Errorf("scan subscription: %w", err)
	}
	sub.CreatedAt = parseTime(createdAt)
	return &sub, nil
}

func scanFire(row scannable) (*core.SubscriptionFire, error) {
	var fire core.SubscriptionFire
	var firedAt string
	if err := row.Scan(&fire.ID, &fire.SubscriptionID, &fire.TriggeringDocID, &fire.TriggeringRef, &fire.MatchedTag, &fire.Status, &firedAt); err != nil {
		return nil, fmt.Errorf("scan subscription fire: %w", err)
	}
	fire.FiredAt = parseTime(firedAt)
	return &fire, nil
}
