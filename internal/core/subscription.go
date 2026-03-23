package core

import "time"

// SubscriptionStatus represents the lifecycle state of a subscription.
type SubscriptionStatus string

const (
	SubActive   SubscriptionStatus = "active"
	SubArchived SubscriptionStatus = "archived"
)

// FireStatus represents the state of a subscription fire.
type FireStatus string

const (
	FirePending   FireStatus = "pending"
	FireDismissed FireStatus = "dismissed"
)

// Subscription represents a tag-based watch on document sections.
// When a section tagged with TagExpression changes, the subscription fires.
type Subscription struct {
	ID            string
	SubscriberRef string             // the section @ref that subscribes
	TagExpression string             // tag to watch for
	Status        SubscriptionStatus // "active" | "archived"
	CreatedBy     string             // principal ID
	CreatedAt     time.Time
}

// NewSubscription creates a new active subscription with an auto-generated ID.
func NewSubscription(subscriberRef, tagExpr, createdBy string) *Subscription {
	return &Subscription{
		ID:            NewID().String(),
		SubscriberRef: subscriberRef,
		TagExpression: tagExpr,
		Status:        SubActive,
		CreatedBy:     createdBy,
		CreatedAt:     time.Now(),
	}
}

// SubscriptionFire records that a subscription matched a triggering change.
type SubscriptionFire struct {
	ID              string
	SubscriptionID  string
	TriggeringDocID string
	TriggeringRef   string
	MatchedTag      string
	Status          FireStatus
	FiredAt         time.Time
}
