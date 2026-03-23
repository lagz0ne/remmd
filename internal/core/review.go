package core

import (
	"context"
	"log/slog"
	"time"
)

// LinkFinder finds and updates link state.
type LinkFinder interface {
	FindLinkByID(ctx context.Context, id string) (*Link, error)
	UpdateLinkState(ctx context.Context, id string, state LinkState) error
}

// ThreadWriter appends entries to a link's discussion thread.
type ThreadWriter interface {
	AddThreadEntry(ctx context.Context, linkID string, entryType EntryType, principalID, body string) error
}

// SnapshotComputer computes the current agreement snapshot for a link.
type SnapshotComputer interface {
	ComputeSnapshot(ctx context.Context, linkID string) (*AgreementSnapshot, error)
}

// ReviewService orchestrates human review actions on links.
type ReviewService struct {
	links     LinkFinder
	threads   ThreadWriter
	snapshots SnapshotComputer
}

// NewReviewService creates a ReviewService with injected dependencies.
func NewReviewService(lf LinkFinder, tw ThreadWriter, sc SnapshotComputer) *ReviewService {
	return &ReviewService{links: lf, threads: tw, snapshots: sc}
}

// Approve transitions a link to aligned after verifying the principal is
// human and the context hash matches the current snapshot.
func (rs *ReviewService) Approve(ctx context.Context, principal Principal, linkID string, contextHash string) error {
	start := time.Now()

	slog.DebugContext(ctx, "attempting link approval",
		"op", "review.approve",
		"entity_type", "link",
		"entity_id", linkID,
		"principal_id", principal.ID,
		"principal_type", string(principal.Type),
		"context_hash", contextHash,
	)

	if err := principal.RequireHuman("approve"); err != nil {
		slog.WarnContext(ctx, "link approval denied: requires human principal",
			"op", "review.approve",
			"entity_type", "link",
			"entity_id", linkID,
			"principal_id", principal.ID,
			"guard", "human_only",
			"principal_type", string(principal.Type),
		)
		return err
	}

	snap, err := rs.snapshots.ComputeSnapshot(ctx, linkID)
	if err != nil {
		slog.ErrorContext(ctx, "link approval failed: snapshot computation error",
			"op", "review.approve",
			"entity_type", "link",
			"entity_id", linkID,
			"outcome", "err",
			"error", err.Error(),
			"duration_ms", time.Since(start).Milliseconds(),
		)
		return err
	}

	currentHash := snap.Hash()
	if contextHash != currentHash {
		slog.WarnContext(ctx, "link approval rejected: stale context",
			"op", "review.approve",
			"entity_type", "link",
			"entity_id", linkID,
			"principal_id", principal.ID,
			"guard", "stale_context",
			"expected_hash", contextHash,
			"actual_hash", currentHash,
		)
		return ErrStaleContext{LinkID: linkID, Expected: contextHash, Actual: currentHash}
	}

	if err := rs.links.UpdateLinkState(ctx, linkID, LinkAligned); err != nil {
		slog.ErrorContext(ctx, "link approval failed: state update error",
			"op", "review.approve",
			"entity_type", "link",
			"entity_id", linkID,
			"outcome", "err",
			"error", err.Error(),
			"duration_ms", time.Since(start).Milliseconds(),
		)
		return err
	}

	_ = rs.threads.AddThreadEntry(ctx, linkID, EntrySystem, principal.ID, "approved")

	slog.InfoContext(ctx, "link approved",
		"op", "review.approve",
		"entity_type", "link",
		"entity_id", linkID,
		"principal_id", principal.ID,
		"outcome", "ok",
		"duration_ms", time.Since(start).Milliseconds(),
	)
	return nil
}

// Reaffirm records the author's reaffirmation after their own content change.
// The link stays stale, waiting for the counterparty to review.
func (rs *ReviewService) Reaffirm(ctx context.Context, principal Principal, linkID string) error {
	start := time.Now()

	slog.DebugContext(ctx, "attempting link reaffirmation",
		"op", "review.reaffirm",
		"entity_type", "link",
		"entity_id", linkID,
		"principal_id", principal.ID,
		"principal_type", string(principal.Type),
	)

	if err := principal.RequireHuman("reaffirm"); err != nil {
		slog.WarnContext(ctx, "link reaffirmation denied: requires human principal",
			"op", "review.reaffirm",
			"entity_type", "link",
			"entity_id", linkID,
			"principal_id", principal.ID,
			"guard", "human_only",
			"principal_type", string(principal.Type),
		)
		return err
	}

	if err := rs.links.UpdateLinkState(ctx, linkID, LinkStale); err != nil {
		slog.ErrorContext(ctx, "link reaffirmation failed: state update error",
			"op", "review.reaffirm",
			"entity_type", "link",
			"entity_id", linkID,
			"outcome", "err",
			"error", err.Error(),
			"duration_ms", time.Since(start).Milliseconds(),
		)
		return err
	}

	_ = rs.threads.AddThreadEntry(ctx, linkID, EntrySystem, principal.ID, "reaffirmed")

	slog.InfoContext(ctx, "link reaffirmed",
		"op", "review.reaffirm",
		"entity_type", "link",
		"entity_id", linkID,
		"principal_id", principal.ID,
		"outcome", "ok",
		"duration_ms", time.Since(start).Milliseconds(),
	)
	return nil
}

// Withdraw archives a link immediately, recording the reason in the thread.
func (rs *ReviewService) Withdraw(ctx context.Context, principal Principal, linkID string, reason string) error {
	start := time.Now()

	slog.DebugContext(ctx, "attempting link withdrawal",
		"op", "review.withdraw",
		"entity_type", "link",
		"entity_id", linkID,
		"principal_id", principal.ID,
		"principal_type", string(principal.Type),
		"reason", reason,
	)

	if err := principal.RequireHuman("withdraw"); err != nil {
		slog.WarnContext(ctx, "link withdrawal denied: requires human principal",
			"op", "review.withdraw",
			"entity_type", "link",
			"entity_id", linkID,
			"principal_id", principal.ID,
			"guard", "human_only",
			"principal_type", string(principal.Type),
		)
		return err
	}

	if err := rs.links.UpdateLinkState(ctx, linkID, LinkArchived); err != nil {
		slog.ErrorContext(ctx, "link withdrawal failed: state update error",
			"op", "review.withdraw",
			"entity_type", "link",
			"entity_id", linkID,
			"outcome", "err",
			"error", err.Error(),
			"duration_ms", time.Since(start).Milliseconds(),
		)
		return err
	}

	_ = rs.threads.AddThreadEntry(ctx, linkID, EntrySystem, principal.ID, reason)

	slog.InfoContext(ctx, "link withdrawn",
		"op", "review.withdraw",
		"entity_type", "link",
		"entity_id", linkID,
		"principal_id", principal.ID,
		"outcome", "ok",
		"reason", reason,
		"duration_ms", time.Since(start).Milliseconds(),
	)
	return nil
}
