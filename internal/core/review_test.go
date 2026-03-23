package core_test

import (
	"context"
	"errors"
	"testing"

	"github.com/lagz0ne/remmd/internal/core"
)

// --- mocks ---

type mockLinkFinder struct {
	link *core.Link
	err  error

	updatedID    string
	updatedState core.LinkState
	updateErr    error
}

func (m *mockLinkFinder) FindLinkByID(_ context.Context, id string) (*core.Link, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.link == nil || m.link.ID != id {
		return nil, core.ErrNotFound{Entity: "link", ID: id}
	}
	return m.link, nil
}

func (m *mockLinkFinder) UpdateLinkState(_ context.Context, id string, state core.LinkState) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.updatedID = id
	m.updatedState = state
	return nil
}

type mockThreadWriter struct {
	entries []threadWriteCall
}

type threadWriteCall struct {
	linkID      string
	entryType   core.EntryType
	principalID string
	body        string
}

func (m *mockThreadWriter) AddThreadEntry(_ context.Context, linkID string, entryType core.EntryType, principalID, body string) error {
	m.entries = append(m.entries, threadWriteCall{linkID, entryType, principalID, body})
	return nil
}

type mockSnapshotComputer struct {
	snapshot *core.AgreementSnapshot
	err      error
}

func (m *mockSnapshotComputer) ComputeSnapshot(_ context.Context, linkID string) (*core.AgreementSnapshot, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.snapshot, nil
}

// --- tests ---

func TestReviewService_Approve_HumanPrincipal_MatchingContext(t *testing.T) {
	t.Parallel()
	link := &core.Link{ID: "l1", State: core.LinkStale}
	snap := &core.AgreementSnapshot{
		LinkID:             "l1",
		LeftContentHashes:  []string{"aaa"},
		RightContentHashes: []string{"bbb"},
	}
	lf := &mockLinkFinder{link: link}
	tw := &mockThreadWriter{}
	sc := &mockSnapshotComputer{snapshot: snap}

	svc := core.NewReviewService(lf, tw, sc)
	principal := core.Principal{ID: "u1", Type: core.PrincipalHuman, Name: "Alice"}

	err := svc.Approve(context.Background(), principal, "l1", snap.Hash())
	if err != nil {
		t.Fatalf("Approve() = %v, want nil", err)
	}
	if lf.updatedState != core.LinkAligned {
		t.Errorf("state = %q, want %q", lf.updatedState, core.LinkAligned)
	}
}

func TestReviewService_Approve_ServicePrincipal_ReturnsUnauthorized(t *testing.T) {
	t.Parallel()
	svc := core.NewReviewService(&mockLinkFinder{}, &mockThreadWriter{}, &mockSnapshotComputer{})
	principal := core.Principal{ID: "svc1", Type: core.PrincipalService, Name: "bot"}

	err := svc.Approve(context.Background(), principal, "l1", "somehash")
	if err == nil {
		t.Fatal("Approve() = nil, want ErrUnauthorized")
	}
	var unauth core.ErrUnauthorized
	if !errors.As(err, &unauth) {
		t.Fatalf("error type = %T, want core.ErrUnauthorized", err)
	}
}

func TestReviewService_Approve_StaleContextHash(t *testing.T) {
	t.Parallel()
	link := &core.Link{ID: "l1", State: core.LinkStale}
	snap := &core.AgreementSnapshot{
		LinkID:             "l1",
		LeftContentHashes:  []string{"aaa"},
		RightContentHashes: []string{"bbb"},
	}
	lf := &mockLinkFinder{link: link}
	tw := &mockThreadWriter{}
	sc := &mockSnapshotComputer{snapshot: snap}

	svc := core.NewReviewService(lf, tw, sc)
	principal := core.Principal{ID: "u1", Type: core.PrincipalHuman, Name: "Alice"}

	err := svc.Approve(context.Background(), principal, "l1", "wronghash")
	if err == nil {
		t.Fatal("Approve() = nil, want ErrStaleContext")
	}
	var stale core.ErrStaleContext
	if !errors.As(err, &stale) {
		t.Fatalf("error type = %T, want core.ErrStaleContext", err)
	}
	if stale.Expected != "wronghash" {
		t.Errorf("Expected = %q, want %q", stale.Expected, "wronghash")
	}
	if stale.Actual != snap.Hash() {
		t.Errorf("Actual = %q, want %q", stale.Actual, snap.Hash())
	}
}

func TestReviewService_Approve_NonExistentLink(t *testing.T) {
	t.Parallel()
	lf := &mockLinkFinder{err: core.ErrNotFound{Entity: "link", ID: "l999"}}
	sc := &mockSnapshotComputer{err: core.ErrNotFound{Entity: "link", ID: "l999"}}
	svc := core.NewReviewService(lf, &mockThreadWriter{}, sc)
	principal := core.Principal{ID: "u1", Type: core.PrincipalHuman, Name: "Alice"}

	err := svc.Approve(context.Background(), principal, "l999", "hash")
	if err == nil {
		t.Fatal("Approve() = nil, want ErrNotFound")
	}
	var notFound core.ErrNotFound
	if !errors.As(err, &notFound) {
		t.Fatalf("error type = %T, want core.ErrNotFound", err)
	}
}

func TestReviewService_Reaffirm_HumanPrincipal(t *testing.T) {
	t.Parallel()
	link := &core.Link{ID: "l1", State: core.LinkStale}
	lf := &mockLinkFinder{link: link}
	tw := &mockThreadWriter{}
	sc := &mockSnapshotComputer{}

	svc := core.NewReviewService(lf, tw, sc)
	principal := core.Principal{ID: "u1", Type: core.PrincipalHuman, Name: "Alice"}

	err := svc.Reaffirm(context.Background(), principal, "l1")
	if err != nil {
		t.Fatalf("Reaffirm() = %v, want nil", err)
	}
	// Reaffirm transitions to stale (waiting on counterparty)
	if lf.updatedState != core.LinkStale {
		t.Errorf("state = %q, want %q", lf.updatedState, core.LinkStale)
	}
}

func TestReviewService_Reaffirm_ServicePrincipal_ReturnsUnauthorized(t *testing.T) {
	t.Parallel()
	svc := core.NewReviewService(&mockLinkFinder{}, &mockThreadWriter{}, &mockSnapshotComputer{})
	principal := core.Principal{ID: "svc1", Type: core.PrincipalService, Name: "bot"}

	err := svc.Reaffirm(context.Background(), principal, "l1")
	if err == nil {
		t.Fatal("Reaffirm() = nil, want ErrUnauthorized")
	}
	var unauth core.ErrUnauthorized
	if !errors.As(err, &unauth) {
		t.Fatalf("error type = %T, want core.ErrUnauthorized", err)
	}
}

func TestReviewService_Withdraw_HumanPrincipal(t *testing.T) {
	t.Parallel()
	link := &core.Link{ID: "l1", State: core.LinkAligned}
	lf := &mockLinkFinder{link: link}
	tw := &mockThreadWriter{}
	sc := &mockSnapshotComputer{}

	svc := core.NewReviewService(lf, tw, sc)
	principal := core.Principal{ID: "u1", Type: core.PrincipalHuman, Name: "Alice"}

	err := svc.Withdraw(context.Background(), principal, "l1", "no longer needed")
	if err != nil {
		t.Fatalf("Withdraw() = %v, want nil", err)
	}
	if lf.updatedState != core.LinkArchived {
		t.Errorf("state = %q, want %q", lf.updatedState, core.LinkArchived)
	}
	if len(tw.entries) == 0 {
		t.Fatal("expected thread entry to be added")
	}
	if tw.entries[0].body != "no longer needed" {
		t.Errorf("thread body = %q, want %q", tw.entries[0].body, "no longer needed")
	}
}

func TestReviewService_Withdraw_ServicePrincipal_ReturnsUnauthorized(t *testing.T) {
	t.Parallel()
	svc := core.NewReviewService(&mockLinkFinder{}, &mockThreadWriter{}, &mockSnapshotComputer{})
	principal := core.Principal{ID: "svc1", Type: core.PrincipalService, Name: "bot"}

	err := svc.Withdraw(context.Background(), principal, "l1", "reason")
	if err == nil {
		t.Fatal("Withdraw() = nil, want ErrUnauthorized")
	}
	var unauth core.ErrUnauthorized
	if !errors.As(err, &unauth) {
		t.Fatalf("error type = %T, want core.ErrUnauthorized", err)
	}
}
