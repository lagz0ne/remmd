package store_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/lagz0ne/remmd/internal/core"
	"github.com/lagz0ne/remmd/internal/store"
)

func setupLinkTestDB(t *testing.T) *sql.DB {
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

// createTestSection creates a document and section in the DB and returns the section.
func createTestSection(t *testing.T, db *sql.DB, docTitle, ref, content string) *core.Section {
	t.Helper()
	docRepo := store.NewDocumentRepo(db)
	ctx := context.Background()

	doc := core.NewDocument(docTitle, "test-user")
	if err := docRepo.CreateDocument(ctx, doc); err != nil {
		t.Fatalf("CreateDocument: %v", err)
	}

	sections := core.Parse(doc.ID, content, 0)
	if len(sections) == 0 {
		t.Fatalf("no sections parsed from content")
	}
	for i := range sections {
		if err := docRepo.CreateSection(ctx, &sections[i]); err != nil {
			t.Fatalf("CreateSection: %v", err)
		}
	}

	// Find the section matching the given ref
	sec, err := docRepo.FindSectionByRef(ctx, doc.ID, ref)
	if err != nil {
		t.Fatalf("FindSectionByRef(%s, %s): %v", doc.ID, ref, err)
	}
	return sec
}

// makeTestLink creates two real sections and links them.
func makeTestLink(t *testing.T, db *sql.DB) *core.Link {
	t.Helper()
	secA := createTestSection(t, db, "Doc A "+core.NewID().String(), "@a1", "# Section A")
	secB := createTestSection(t, db, "Doc B "+core.NewID().String(), "@a1", "# Section B")
	return core.NewLink(
		[]string{secA.ID},
		[]string{secB.ID},
		core.RelImplements,
		core.Rationale{Claim: "impl matches spec", Scope: "all endpoints", Exclusions: ""},
		"user-1",
	)
}

func TestCreateLink_And_FindByID(t *testing.T) {
	t.Parallel()
	db := setupLinkTestDB(t)
	repo := store.NewLinkRepo(db)
	ctx := context.Background()

	link := makeTestLink(t, db)
	if err := repo.CreateLink(ctx, link); err != nil {
		t.Fatalf("CreateLink: %v", err)
	}

	got, err := repo.FindLinkByID(ctx, link.ID)
	if err != nil {
		t.Fatalf("FindLinkByID: %v", err)
	}
	if got.ID != link.ID {
		t.Errorf("ID = %q, want %q", got.ID, link.ID)
	}
	if got.RelationshipType != core.RelImplements {
		t.Errorf("RelationshipType = %q, want %q", got.RelationshipType, core.RelImplements)
	}
	if got.State != core.LinkPending {
		t.Errorf("State = %q, want %q", got.State, core.LinkPending)
	}
	if got.ProposerID != "user-1" {
		t.Errorf("ProposerID = %q, want %q", got.ProposerID, "user-1")
	}
	if got.Rationale.Claim != "impl matches spec" {
		t.Errorf("Rationale.Claim = %q, want %q", got.Rationale.Claim, "impl matches spec")
	}
	if len(got.LeftSectionIDs) != 1 || got.LeftSectionIDs[0] != link.LeftSectionIDs[0] {
		t.Errorf("LeftSectionIDs = %v, want %v", got.LeftSectionIDs, link.LeftSectionIDs)
	}
	if len(got.RightSectionIDs) != 1 || got.RightSectionIDs[0] != link.RightSectionIDs[0] {
		t.Errorf("RightSectionIDs = %v, want %v", got.RightSectionIDs, link.RightSectionIDs)
	}
	if got.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
}

func TestFindLinkByID_NotFound(t *testing.T) {
	t.Parallel()
	db := setupLinkTestDB(t)
	repo := store.NewLinkRepo(db)
	ctx := context.Background()

	_, err := repo.FindLinkByID(ctx, "does-not-exist")
	if err == nil {
		t.Fatal("expected error for non-existent link")
	}
	if !errors.Is(err, core.ErrNotFound{}) {
		t.Errorf("expected ErrNotFound, got %T: %v", err, err)
	}
}

func TestListLinks_ReturnsAll(t *testing.T) {
	t.Parallel()
	db := setupLinkTestDB(t)
	repo := store.NewLinkRepo(db)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		link := makeTestLink(t, db)
		if err := repo.CreateLink(ctx, link); err != nil {
			t.Fatalf("CreateLink #%d: %v", i, err)
		}
	}

	links, err := repo.ListLinks(ctx, "")
	if err != nil {
		t.Fatalf("ListLinks: %v", err)
	}
	if len(links) != 3 {
		t.Fatalf("expected 3 links, got %d", len(links))
	}
}

func TestListLinks_StateFilter(t *testing.T) {
	t.Parallel()
	db := setupLinkTestDB(t)
	repo := store.NewLinkRepo(db)
	ctx := context.Background()

	// Create two pending links
	for i := 0; i < 2; i++ {
		link := makeTestLink(t, db)
		if err := repo.CreateLink(ctx, link); err != nil {
			t.Fatalf("CreateLink: %v", err)
		}
	}

	// Create one and move to aligned
	link := makeTestLink(t, db)
	if err := repo.CreateLink(ctx, link); err != nil {
		t.Fatalf("CreateLink: %v", err)
	}
	if err := repo.UpdateLinkState(ctx, link.ID, core.LinkAligned); err != nil {
		t.Fatalf("UpdateLinkState: %v", err)
	}

	// Filter by pending
	pending, err := repo.ListLinks(ctx, "pending")
	if err != nil {
		t.Fatalf("ListLinks(pending): %v", err)
	}
	if len(pending) != 2 {
		t.Errorf("expected 2 pending links, got %d", len(pending))
	}

	// Filter by aligned
	aligned, err := repo.ListLinks(ctx, "aligned")
	if err != nil {
		t.Fatalf("ListLinks(aligned): %v", err)
	}
	if len(aligned) != 1 {
		t.Errorf("expected 1 aligned link, got %d", len(aligned))
	}
}

func TestUpdateLinkState(t *testing.T) {
	t.Parallel()
	db := setupLinkTestDB(t)
	repo := store.NewLinkRepo(db)
	ctx := context.Background()

	link := makeTestLink(t, db)
	if err := repo.CreateLink(ctx, link); err != nil {
		t.Fatalf("CreateLink: %v", err)
	}

	if err := repo.UpdateLinkState(ctx, link.ID, core.LinkAligned); err != nil {
		t.Fatalf("UpdateLinkState: %v", err)
	}

	got, err := repo.FindLinkByID(ctx, link.ID)
	if err != nil {
		t.Fatalf("FindLinkByID: %v", err)
	}
	if got.State != core.LinkAligned {
		t.Errorf("State = %q, want %q", got.State, core.LinkAligned)
	}
}

func TestAddThreadEntry_And_GetThread(t *testing.T) {
	t.Parallel()
	db := setupLinkTestDB(t)
	repo := store.NewLinkRepo(db)
	ctx := context.Background()

	link := makeTestLink(t, db)
	if err := repo.CreateLink(ctx, link); err != nil {
		t.Fatalf("CreateLink: %v", err)
	}

	if err := repo.AddThreadEntry(ctx, link.ID, core.EntryComment, "user-1", "Looks good to me"); err != nil {
		t.Fatalf("AddThreadEntry: %v", err)
	}

	entries, err := repo.GetThread(ctx, link.ID)
	if err != nil {
		t.Fatalf("GetThread: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Body != "Looks good to me" {
		t.Errorf("Body = %q, want %q", entries[0].Body, "Looks good to me")
	}
	if entries[0].LinkID != link.ID {
		t.Errorf("LinkID = %q, want %q", entries[0].LinkID, link.ID)
	}
	if entries[0].Type != core.EntryComment {
		t.Errorf("Type = %q, want %q", entries[0].Type, core.EntryComment)
	}
}

func TestLinksContainingSection(t *testing.T) {
	t.Parallel()
	db := setupLinkTestDB(t)
	repo := store.NewLinkRepo(db)
	ctx := context.Background()

	// Create real sections with globally unique IDs
	secA1 := createTestSection(t, db, "Doc A", "@a1", "# Section A1")
	secB1 := createTestSection(t, db, "Doc B", "@a1", "# Section B1") // Same @a1 ref, different doc!
	secC1 := createTestSection(t, db, "Doc C", "@a1", "# Section C1") // Yet another @a1

	// link1: secA1 -> secB1 (using section IDs, not refs)
	link1 := core.NewLink(
		[]string{secA1.ID},
		[]string{secB1.ID},
		core.RelImplements,
		core.Rationale{Claim: "impl matches spec", Scope: "all endpoints", Exclusions: ""},
		"user-1",
	)
	if err := repo.CreateLink(ctx, link1); err != nil {
		t.Fatalf("CreateLink link1: %v", err)
	}

	// link2: secA1 -> secC1
	link2 := core.NewLink(
		[]string{secA1.ID},
		[]string{secC1.ID},
		core.RelTests,
		core.Rationale{Claim: "test coverage", Scope: "unit", Exclusions: ""},
		"user-2",
	)
	if err := repo.CreateLink(ctx, link2); err != nil {
		t.Fatalf("CreateLink link2: %v", err)
	}

	// secA1.ID should appear in both links
	links, err := repo.LinksContainingSection(ctx, secA1.ID)
	if err != nil {
		t.Fatalf("LinksContainingSection: %v", err)
	}
	if len(links) != 2 {
		t.Fatalf("expected 2 links containing secA1.ID, got %d", len(links))
	}

	// secB1.ID should appear in only link1
	links, err = repo.LinksContainingSection(ctx, secB1.ID)
	if err != nil {
		t.Fatalf("LinksContainingSection: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("expected 1 link containing secB1.ID, got %d", len(links))
	}

	// secC1.ID should appear in only link2
	links, err = repo.LinksContainingSection(ctx, secC1.ID)
	if err != nil {
		t.Fatalf("LinksContainingSection: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("expected 1 link containing secC1.ID, got %d", len(links))
	}

	// Verify the link stores section IDs, not refs — the key bug test
	got, err := repo.FindLinkByID(ctx, link1.ID)
	if err != nil {
		t.Fatalf("FindLinkByID: %v", err)
	}
	// LeftSectionIDs should contain the actual section ULID, not "@a1"
	if len(got.LeftSectionIDs) != 1 || got.LeftSectionIDs[0] != secA1.ID {
		t.Errorf("LeftSectionIDs = %v, want [%s]", got.LeftSectionIDs, secA1.ID)
	}
	if len(got.RightSectionIDs) != 1 || got.RightSectionIDs[0] != secB1.ID {
		t.Errorf("RightSectionIDs = %v, want [%s]", got.RightSectionIDs, secB1.ID)
	}
}

func TestLinksContainingSection_AmbiguousRefs_Disambiguated(t *testing.T) {
	t.Parallel()
	db := setupLinkTestDB(t)
	repo := store.NewLinkRepo(db)
	ctx := context.Background()

	// THE BUG: Two documents each have a section with ref @a1.
	// Links should be able to distinguish them because we store section IDs.
	docA_secA1 := createTestSection(t, db, "Doc Alpha", "@a1", "# Alpha Section")
	docB_secA1 := createTestSection(t, db, "Doc Beta", "@a1", "# Beta Section")

	// They must have different IDs even though same ref
	if docA_secA1.ID == docB_secA1.ID {
		t.Fatal("two sections with same ref in different docs should have different IDs")
	}

	// Link from Doc Alpha's @a1 to Doc Beta's @a1
	link := core.NewLink(
		[]string{docA_secA1.ID},
		[]string{docB_secA1.ID},
		core.RelAgreesWith,
		core.Rationale{Claim: "aligned", Scope: "full", Exclusions: ""},
		"user-1",
	)
	if err := repo.CreateLink(ctx, link); err != nil {
		t.Fatalf("CreateLink: %v", err)
	}

	// Query by Doc Alpha's section ID — should find the link
	links, err := repo.LinksContainingSection(ctx, docA_secA1.ID)
	if err != nil {
		t.Fatalf("LinksContainingSection(alpha): %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("expected 1 link for alpha's @a1, got %d", len(links))
	}

	// Query by Doc Beta's section ID — should also find the link
	links, err = repo.LinksContainingSection(ctx, docB_secA1.ID)
	if err != nil {
		t.Fatalf("LinksContainingSection(beta): %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("expected 1 link for beta's @a1, got %d", len(links))
	}
}
