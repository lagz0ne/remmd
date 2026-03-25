package store

import (
	"context"
	"fmt"

	"github.com/lagz0ne/remmd/internal/core"
)

// SnapshotService computes agreement snapshots from the current state of links and sections.
type SnapshotService struct {
	links *LinkRepo
	docs  *DocumentRepo
}

// NewSnapshotService creates a SnapshotService with the given repos.
func NewSnapshotService(links *LinkRepo, docs *DocumentRepo) *SnapshotService {
	return &SnapshotService{links: links, docs: docs}
}

// ComputeSnapshot builds an AgreementSnapshot for the given link by reading
// the current content hashes of all sections on both sides.
func (s *SnapshotService) ComputeSnapshot(ctx context.Context, linkID string) (*core.AgreementSnapshot, error) {
	link, err := s.links.FindLinkByID(ctx, linkID)
	if err != nil {
		return nil, fmt.Errorf("find link for snapshot: %w", err)
	}

	leftHashes := make([]string, 0, len(link.LeftSectionIDs))
	for _, sid := range link.LeftSectionIDs {
		sec, err := s.docs.FindSectionByID(ctx, sid)
		if err != nil {
			return nil, fmt.Errorf("find left section %s: %w", sid, err)
		}
		leftHashes = append(leftHashes, sec.ContentHash)
	}

	rightHashes := make([]string, 0, len(link.RightSectionIDs))
	for _, sid := range link.RightSectionIDs {
		sec, err := s.docs.FindSectionByID(ctx, sid)
		if err != nil {
			return nil, fmt.Errorf("find right section %s: %w", sid, err)
		}
		rightHashes = append(rightHashes, sec.ContentHash)
	}

	return &core.AgreementSnapshot{
		LinkID:             linkID,
		LeftContentHashes:  leftHashes,
		RightContentHashes: rightHashes,
	}, nil
}
