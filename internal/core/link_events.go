package core

import "encoding/json"

// NewLinkProposedEvent records that a link was proposed.
func NewLinkProposedEvent(link *Link) Event {
	payload, _ := json.Marshal(map[string]any{
		"relationship_type":  string(link.RelationshipType),
		"left_section_ids":   link.LeftSectionIDs,
		"right_section_ids":  link.RightSectionIDs,
		"rationale_claim":    link.Rationale.Claim,
	})
	return NewEvent(link.ID, "link", "link.proposed", payload, link.ProposerID)
}

// NewEndpointApprovedEvent records that an endpoint approved the link.
func NewEndpointApprovedEvent(linkID, principalID, side string) Event {
	payload, _ := json.Marshal(map[string]string{
		"side": side,
	})
	return NewEvent(linkID, "link", "endpoint.approved", payload, principalID)
}

// NewLinkAlignedEvent records that a link transitioned to aligned.
func NewLinkAlignedEvent(linkID, principalID string) Event {
	payload, _ := json.Marshal(map[string]string{})
	return NewEvent(linkID, "link", "link.aligned", payload, principalID)
}

// NewLinkStaleEvent records that a link went stale due to upstream change.
func NewLinkStaleEvent(linkID, principalID, reason string) Event {
	payload, _ := json.Marshal(map[string]string{
		"reason": reason,
	})
	return NewEvent(linkID, "link", "link.stale", payload, principalID)
}

// NewLinkBrokenEvent records that a link was marked broken.
func NewLinkBrokenEvent(linkID, principalID, reason string) Event {
	payload, _ := json.Marshal(map[string]string{
		"reason": reason,
	})
	return NewEvent(linkID, "link", "link.broken", payload, principalID)
}

// NewLinkArchivedEvent records that a link was archived.
func NewLinkArchivedEvent(linkID, principalID string) Event {
	payload, _ := json.Marshal(map[string]string{})
	return NewEvent(linkID, "link", "link.archived", payload, principalID)
}

// NewCommentAddedEvent records that a comment was added to a link thread.
func NewCommentAddedEvent(linkID, principalID, body string) Event {
	payload, _ := json.Marshal(map[string]string{
		"body": body,
	})
	return NewEvent(linkID, "link", "comment.added", payload, principalID)
}
