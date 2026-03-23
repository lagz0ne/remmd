package core

import "encoding/json"

// NewDocumentCreatedEvent creates an event recording document creation.
func NewDocumentCreatedEvent(doc *Document) Event {
	payload, _ := json.Marshal(map[string]string{
		"title":    doc.Title,
		"owner_id": doc.OwnerID,
		"status":   string(doc.Status),
	})
	return NewEvent(doc.ID, "document", "document.created", payload, doc.OwnerID)
}

// NewSectionEditedEvent creates an event recording a section content edit.
func NewSectionEditedEvent(sectionID, docID, oldHash, newHash, principalID string) Event {
	payload, _ := json.Marshal(map[string]string{
		"section_id": sectionID,
		"old_hash":   oldHash,
		"new_hash":   newHash,
	})
	return NewEvent(docID, "document", "section.edited", payload, principalID)
}

// NewSectionDeletedEvent creates an event recording a section deletion.
func NewSectionDeletedEvent(sectionID, docID, reason, principalID string) Event {
	payload, _ := json.Marshal(map[string]string{
		"section_id": sectionID,
		"reason":     reason,
	})
	return NewEvent(docID, "document", "section.deleted", payload, principalID)
}

// NewSectionTaggedEvent creates an event recording a tag being applied to a section.
func NewSectionTaggedEvent(sectionID, docID, tag, principalID string) Event {
	payload, _ := json.Marshal(map[string]string{
		"section_id": sectionID,
		"tag":        tag,
	})
	return NewEvent(docID, "document", "section.tagged", payload, principalID)
}
