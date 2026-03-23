package core

import "encoding/json"

// NewSubscriptionCreatedEvent records that a subscription was created.
func NewSubscriptionCreatedEvent(sub *Subscription) Event {
	payload, _ := json.Marshal(map[string]string{
		"subscriber_ref": sub.SubscriberRef,
		"tag_expression": sub.TagExpression,
	})
	return NewEvent(sub.ID, "subscription", "subscription.created", payload, sub.CreatedBy)
}

// NewSubscriptionFiredEvent records that a subscription fired due to a matching change.
func NewSubscriptionFiredEvent(fire *SubscriptionFire, principalID string) Event {
	payload, _ := json.Marshal(map[string]string{
		"fire_id":           fire.ID,
		"triggering_doc_id": fire.TriggeringDocID,
		"triggering_ref":    fire.TriggeringRef,
		"matched_tag":       fire.MatchedTag,
	})
	return NewEvent(fire.SubscriptionID, "subscription", "subscription.fired", payload, principalID)
}

// NewSubscriptionDismissedEvent records that a subscription fire was dismissed.
func NewSubscriptionDismissedEvent(fireID, subscriptionID, principalID string) Event {
	payload, _ := json.Marshal(map[string]string{
		"fire_id": fireID,
	})
	return NewEvent(subscriptionID, "subscription", "subscription.dismissed", payload, principalID)
}
