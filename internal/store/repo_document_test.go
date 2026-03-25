package store_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/lagz0ne/remmd/internal/core"
	"github.com/lagz0ne/remmd/internal/store"
)

func setupDocTestDB(t *testing.T) *sql.DB {
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

func TestCreateDocument_And_FindByID(t *testing.T) {
	t.Parallel()
	db := setupDocTestDB(t)
	repo := store.NewDocumentRepo(db)
	ctx := context.Background()

	doc := &core.Document{
		ID:      core.NewID().String(),
		Title:   "Test Document",
		OwnerID: "user-1",
		Status:  core.DocumentActive,
		Source:  "native",
	}

	if err := repo.CreateDocument(ctx, doc); err != nil {
		t.Fatalf("CreateDocument: %v", err)
	}

	got, err := repo.FindDocumentByID(ctx, doc.ID)
	if err != nil {
		t.Fatalf("FindDocumentByID: %v", err)
	}
	if got.ID != doc.ID {
		t.Errorf("ID = %q, want %q", got.ID, doc.ID)
	}
	if got.Title != doc.Title {
		t.Errorf("Title = %q, want %q", got.Title, doc.Title)
	}
	if got.OwnerID != doc.OwnerID {
		t.Errorf("OwnerID = %q, want %q", got.OwnerID, doc.OwnerID)
	}
	if got.Status != core.DocumentActive {
		t.Errorf("Status = %q, want %q", got.Status, core.DocumentActive)
	}
	if got.Source != "native" {
		t.Errorf("Source = %q, want %q", got.Source, "native")
	}
	if got.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
}

func TestFindDocumentByID_NotFound(t *testing.T) {
	t.Parallel()
	db := setupDocTestDB(t)
	repo := store.NewDocumentRepo(db)
	ctx := context.Background()

	_, err := repo.FindDocumentByID(ctx, "does-not-exist")
	if err == nil {
		t.Fatal("expected error for non-existent document")
	}
	if !errors.Is(err, core.ErrNotFound{}) {
		t.Errorf("expected ErrNotFound, got %T: %v", err, err)
	}
}

func TestCreateSection_And_ListSections(t *testing.T) {
	t.Parallel()
	db := setupDocTestDB(t)
	repo := store.NewDocumentRepo(db)
	ctx := context.Background()

	docID := core.NewID().String()
	doc := &core.Document{ID: docID, Title: "Doc", OwnerID: "user-1", Status: "active", Source: "native"}
	if err := repo.CreateDocument(ctx, doc); err != nil {
		t.Fatalf("CreateDocument: %v", err)
	}

	ref1, _ := core.ParseRef("@a1")
	ref2, _ := core.ParseRef("@b2")

	s1 := &core.Section{
		ID:          core.NewID().String(),
		DocID:       docID,
		Ref:         ref1,
		Type:        core.SectionHeading,
		Title:       "Section A",
		Content:     "Hello",
		ContentHash: core.ContentHash("Hello"),
		Order:       1,
	}
	s2 := &core.Section{
		ID:          core.NewID().String(),
		DocID:       docID,
		Ref:         ref2,
		Type:        core.SectionListItem,
		Title:       "Section B",
		Content:     "World",
		ContentHash: core.ContentHash("World"),
		ParentRef:   &ref1,
		Order:       2,
	}

	if err := repo.CreateSection(ctx, s1); err != nil {
		t.Fatalf("CreateSection s1: %v", err)
	}
	if err := repo.CreateSection(ctx, s2); err != nil {
		t.Fatalf("CreateSection s2: %v", err)
	}

	sections, err := repo.ListSections(ctx, docID)
	if err != nil {
		t.Fatalf("ListSections: %v", err)
	}
	if len(sections) != 2 {
		t.Fatalf("expected 2 sections, got %d", len(sections))
	}

	// Ordered by "order"
	if sections[0].Order != 1 {
		t.Errorf("first section order = %d, want 1", sections[0].Order)
	}
	if sections[1].Order != 2 {
		t.Errorf("second section order = %d, want 2", sections[1].Order)
	}
	if sections[0].Ref.String() != "@a1" {
		t.Errorf("first section ref = %q, want @a1", sections[0].Ref.String())
	}
	if sections[1].ParentRef == nil {
		t.Error("second section should have a parent ref")
	} else if sections[1].ParentRef.String() != "@a1" {
		t.Errorf("second section parent ref = %q, want @a1", sections[1].ParentRef.String())
	}
}

func TestListSections_OrderedByOrderField(t *testing.T) {
	t.Parallel()
	db := setupDocTestDB(t)
	repo := store.NewDocumentRepo(db)
	ctx := context.Background()

	docID := core.NewID().String()
	doc := &core.Document{ID: docID, Title: "Doc", OwnerID: "user-1", Status: "active", Source: "native"}
	if err := repo.CreateDocument(ctx, doc); err != nil {
		t.Fatalf("CreateDocument: %v", err)
	}

	// Insert in reverse order
	refs := []string{"@c3", "@b2", "@a1"}
	for i, r := range refs {
		ref, _ := core.ParseRef(r)
		s := &core.Section{
			ID:          core.NewID().String(),
			DocID:       docID,
			Ref:         ref,
			Type:        core.SectionHeading,
			Content:     r,
			ContentHash: core.ContentHash(r),
			Order:       3 - i, // 3, 2, 1
		}
		if err := repo.CreateSection(ctx, s); err != nil {
			t.Fatalf("CreateSection: %v", err)
		}
	}

	sections, err := repo.ListSections(ctx, docID)
	if err != nil {
		t.Fatalf("ListSections: %v", err)
	}
	// Should be ordered 1, 2, 3
	for i, s := range sections {
		want := i + 1
		if s.Order != want {
			t.Errorf("sections[%d].Order = %d, want %d", i, s.Order, want)
		}
	}
}

func TestUpdateSectionContent_CreatesVersion(t *testing.T) {
	t.Parallel()
	db := setupDocTestDB(t)
	repo := store.NewDocumentRepo(db)
	ctx := context.Background()

	docID := core.NewID().String()
	doc := &core.Document{ID: docID, Title: "Doc", OwnerID: "user-1", Status: "active", Source: "native"}
	if err := repo.CreateDocument(ctx, doc); err != nil {
		t.Fatalf("CreateDocument: %v", err)
	}

	ref, _ := core.ParseRef("@a1")
	sectionID := core.NewID().String()
	s := &core.Section{
		ID:          sectionID,
		DocID:       docID,
		Ref:         ref,
		Type:        core.SectionHeading,
		Content:     "v1 content",
		ContentHash: core.ContentHash("v1 content"),
		Order:       1,
	}
	if err := repo.CreateSection(ctx, s); err != nil {
		t.Fatalf("CreateSection: %v", err)
	}

	newContent := "v2 content"
	newHash := core.ContentHash(newContent)
	if err := repo.UpdateSectionContent(ctx, sectionID, newContent, newHash); err != nil {
		t.Fatalf("UpdateSectionContent: %v", err)
	}

	// Section should have updated content
	sections, err := repo.ListSections(ctx, docID)
	if err != nil {
		t.Fatalf("ListSections: %v", err)
	}
	if len(sections) != 1 {
		t.Fatalf("expected 1 section, got %d", len(sections))
	}
	if sections[0].Content != newContent {
		t.Errorf("Content = %q, want %q", sections[0].Content, newContent)
	}
	if sections[0].ContentHash != newHash {
		t.Errorf("ContentHash = %q, want %q", sections[0].ContentHash, newHash)
	}
}

func TestGetSectionVersions(t *testing.T) {
	t.Parallel()
	db := setupDocTestDB(t)
	repo := store.NewDocumentRepo(db)
	ctx := context.Background()

	docID := core.NewID().String()
	doc := &core.Document{ID: docID, Title: "Doc", OwnerID: "user-1", Status: "active", Source: "native"}
	if err := repo.CreateDocument(ctx, doc); err != nil {
		t.Fatalf("CreateDocument: %v", err)
	}

	ref, _ := core.ParseRef("@a1")
	sectionID := core.NewID().String()
	s := &core.Section{
		ID:          sectionID,
		DocID:       docID,
		Ref:         ref,
		Type:        core.SectionHeading,
		Content:     "original",
		ContentHash: core.ContentHash("original"),
		Order:       1,
	}
	if err := repo.CreateSection(ctx, s); err != nil {
		t.Fatalf("CreateSection: %v", err)
	}

	// Update twice
	for _, c := range []string{"update-1", "update-2"} {
		if err := repo.UpdateSectionContent(ctx, sectionID, c, core.ContentHash(c)); err != nil {
			t.Fatalf("UpdateSectionContent %q: %v", c, err)
		}
	}

	versions, err := repo.GetSectionVersions(ctx, sectionID)
	if err != nil {
		t.Fatalf("GetSectionVersions: %v", err)
	}
	if len(versions) != 2 {
		t.Fatalf("expected 2 versions, got %d", len(versions))
	}
	if versions[0].Version != 1 {
		t.Errorf("versions[0].Version = %d, want 1", versions[0].Version)
	}
	if versions[1].Version != 2 {
		t.Errorf("versions[1].Version = %d, want 2", versions[1].Version)
	}
	if versions[0].Content != "update-1" {
		t.Errorf("versions[0].Content = %q, want %q", versions[0].Content, "update-1")
	}
	if versions[1].Content != "update-2" {
		t.Errorf("versions[1].Content = %q, want %q", versions[1].Content, "update-2")
	}
}

func TestAddTag_And_GetTags(t *testing.T) {
	t.Parallel()
	db := setupDocTestDB(t)
	repo := store.NewDocumentRepo(db)
	ctx := context.Background()

	docID := core.NewID().String()
	doc := &core.Document{ID: docID, Title: "Doc", OwnerID: "user-1", Status: "active", Source: "native"}
	if err := repo.CreateDocument(ctx, doc); err != nil {
		t.Fatalf("CreateDocument: %v", err)
	}

	ref, _ := core.ParseRef("@a1")
	sectionID := core.NewID().String()
	s := &core.Section{
		ID:          sectionID,
		DocID:       docID,
		Ref:         ref,
		Type:        core.SectionHeading,
		Content:     "tagged",
		ContentHash: core.ContentHash("tagged"),
		Order:       1,
	}
	if err := repo.CreateSection(ctx, s); err != nil {
		t.Fatalf("CreateSection: %v", err)
	}

	for _, tag := range []string{"api", "v2", "critical"} {
		if err := repo.AddTag(ctx, sectionID, tag); err != nil {
			t.Fatalf("AddTag %q: %v", tag, err)
		}
	}

	tags, err := repo.GetTags(ctx, sectionID)
	if err != nil {
		t.Fatalf("GetTags: %v", err)
	}
	if len(tags) != 3 {
		t.Fatalf("expected 3 tags, got %d", len(tags))
	}

	// Tags should be sorted
	want := []string{"api", "critical", "v2"}
	for i, tag := range tags {
		if tag != want[i] {
			t.Errorf("tags[%d] = %q, want %q", i, tag, want[i])
		}
	}
}

func TestAddTag_Idempotent(t *testing.T) {
	t.Parallel()
	db := setupDocTestDB(t)
	repo := store.NewDocumentRepo(db)
	ctx := context.Background()

	docID := core.NewID().String()
	doc := &core.Document{ID: docID, Title: "Doc", OwnerID: "user-1", Status: "active", Source: "native"}
	if err := repo.CreateDocument(ctx, doc); err != nil {
		t.Fatalf("CreateDocument: %v", err)
	}

	ref, _ := core.ParseRef("@a1")
	sectionID := core.NewID().String()
	s := &core.Section{
		ID:          sectionID,
		DocID:       docID,
		Ref:         ref,
		Type:        core.SectionHeading,
		Content:     "tagged",
		ContentHash: core.ContentHash("tagged"),
		Order:       1,
	}
	if err := repo.CreateSection(ctx, s); err != nil {
		t.Fatalf("CreateSection: %v", err)
	}

	// Add same tag twice -- should not error
	if err := repo.AddTag(ctx, sectionID, "api"); err != nil {
		t.Fatalf("first AddTag: %v", err)
	}
	if err := repo.AddTag(ctx, sectionID, "api"); err != nil {
		t.Fatalf("second AddTag should be idempotent: %v", err)
	}

	tags, err := repo.GetTags(ctx, sectionID)
	if err != nil {
		t.Fatalf("GetTags: %v", err)
	}
	if len(tags) != 1 {
		t.Fatalf("expected 1 tag after double add, got %d", len(tags))
	}
}

func TestDeleteSection(t *testing.T) {
	t.Parallel()
	db := setupDocTestDB(t)
	repo := store.NewDocumentRepo(db)
	ctx := context.Background()

	docID := core.NewID().String()
	doc := &core.Document{ID: docID, Title: "Doc", OwnerID: "user-1", Status: "active", Source: "native"}
	if err := repo.CreateDocument(ctx, doc); err != nil {
		t.Fatalf("CreateDocument: %v", err)
	}

	ref, _ := core.ParseRef("@a1")
	sectionID := core.NewID().String()
	s := &core.Section{
		ID:          sectionID,
		DocID:       docID,
		Ref:         ref,
		Type:        core.SectionHeading,
		Content:     "to delete",
		ContentHash: core.ContentHash("to delete"),
		Order:       1,
	}
	if err := repo.CreateSection(ctx, s); err != nil {
		t.Fatalf("CreateSection: %v", err)
	}

	if err := repo.DeleteSection(ctx, sectionID); err != nil {
		t.Fatalf("DeleteSection: %v", err)
	}

	sections, err := repo.ListSections(ctx, docID)
	if err != nil {
		t.Fatalf("ListSections: %v", err)
	}
	if len(sections) != 0 {
		t.Fatalf("expected 0 sections after delete, got %d", len(sections))
	}
}

func TestListDocuments(t *testing.T) {
	t.Parallel()
	db := setupDocTestDB(t)
	repo := store.NewDocumentRepo(db)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		doc := &core.Document{
			ID:      core.NewID().String(),
			Title:   "Doc",
			OwnerID: "user-1",
			Status:  core.DocumentActive,
			Source:  "native",
		}
		if err := repo.CreateDocument(ctx, doc); err != nil {
			t.Fatalf("CreateDocument #%d: %v", i, err)
		}
	}

	docs, err := repo.ListDocuments(ctx)
	if err != nil {
		t.Fatalf("ListDocuments: %v", err)
	}
	if len(docs) != 3 {
		t.Fatalf("expected 3 documents, got %d", len(docs))
	}
}

func TestFindSectionByRef(t *testing.T) {
	t.Parallel()
	db := setupDocTestDB(t)
	repo := store.NewDocumentRepo(db)
	ctx := context.Background()

	docID := core.NewID().String()
	doc := &core.Document{ID: docID, Title: "Doc", OwnerID: "user-1", Status: "active", Source: "native"}
	if err := repo.CreateDocument(ctx, doc); err != nil {
		t.Fatalf("CreateDocument: %v", err)
	}

	ref, _ := core.ParseRef("@a1")
	sectionID := core.NewID().String()
	s := &core.Section{
		ID:          sectionID,
		DocID:       docID,
		Ref:         ref,
		Type:        core.SectionHeading,
		Content:     "findme",
		ContentHash: core.ContentHash("findme"),
		Order:       1,
	}
	if err := repo.CreateSection(ctx, s); err != nil {
		t.Fatalf("CreateSection: %v", err)
	}

	got, err := repo.FindSectionByRef(ctx, docID, "@a1")
	if err != nil {
		t.Fatalf("FindSectionByRef: %v", err)
	}
	if got.ID != sectionID {
		t.Errorf("ID = %q, want %q", got.ID, sectionID)
	}
	if got.Content != "findme" {
		t.Errorf("Content = %q, want %q", got.Content, "findme")
	}
}

func TestFindSectionByRefGlobal(t *testing.T) {
	t.Parallel()
	db := setupDocTestDB(t)
	repo := store.NewDocumentRepo(db)
	ctx := context.Background()

	// Create two documents with sections
	doc1ID := core.NewID().String()
	doc1 := &core.Document{ID: doc1ID, Title: "Doc1", OwnerID: "user-1", Status: "active", Source: "native"}
	if err := repo.CreateDocument(ctx, doc1); err != nil {
		t.Fatalf("CreateDocument doc1: %v", err)
	}

	doc2ID := core.NewID().String()
	doc2 := &core.Document{ID: doc2ID, Title: "Doc2", OwnerID: "user-1", Status: "active", Source: "native"}
	if err := repo.CreateDocument(ctx, doc2); err != nil {
		t.Fatalf("CreateDocument doc2: %v", err)
	}

	ref1, _ := core.ParseRef("@x1")
	sec1ID := core.NewID().String()
	s1 := &core.Section{
		ID: sec1ID, DocID: doc1ID, Ref: ref1, Type: core.SectionHeading,
		Content: "in doc1", ContentHash: core.ContentHash("in doc1"), Order: 1,
	}
	if err := repo.CreateSection(ctx, s1); err != nil {
		t.Fatalf("CreateSection s1: %v", err)
	}

	ref2, _ := core.ParseRef("@y2")
	sec2ID := core.NewID().String()
	s2 := &core.Section{
		ID: sec2ID, DocID: doc2ID, Ref: ref2, Type: core.SectionListItem,
		Content: "in doc2", ContentHash: core.ContentHash("in doc2"), Order: 1,
	}
	if err := repo.CreateSection(ctx, s2); err != nil {
		t.Fatalf("CreateSection s2: %v", err)
	}

	// Find section in doc1
	got, gotDocID, err := repo.FindSectionByRefGlobal(ctx, "@x1")
	if err != nil {
		t.Fatalf("FindSectionByRefGlobal(@x1): %v", err)
	}
	if got.ID != sec1ID {
		t.Errorf("ID = %q, want %q", got.ID, sec1ID)
	}
	if gotDocID != doc1ID {
		t.Errorf("DocID = %q, want %q", gotDocID, doc1ID)
	}

	// Find section in doc2
	got2, gotDocID2, err := repo.FindSectionByRefGlobal(ctx, "@y2")
	if err != nil {
		t.Fatalf("FindSectionByRefGlobal(@y2): %v", err)
	}
	if got2.ID != sec2ID {
		t.Errorf("ID = %q, want %q", got2.ID, sec2ID)
	}
	if gotDocID2 != doc2ID {
		t.Errorf("DocID = %q, want %q", gotDocID2, doc2ID)
	}

	// Not found
	_, _, err = repo.FindSectionByRefGlobal(ctx, "@zz99")
	if err == nil {
		t.Fatal("expected error for non-existent ref")
	}
	if !errors.Is(err, core.ErrNotFound{}) {
		t.Errorf("expected ErrNotFound, got %T: %v", err, err)
	}
}

// --- External content support (RED tests — won't compile until implementation exists) ---

func TestCreateSection_External(t *testing.T) {
	t.Parallel()
	db := setupDocTestDB(t)
	repo := store.NewDocumentRepo(db)
	ctx := context.Background()

	docID := core.NewID().String()
	doc := &core.Document{ID: docID, Title: "Doc", OwnerID: "user-1", Status: "active", Source: "native"}
	if err := repo.CreateDocument(ctx, doc); err != nil {
		t.Fatalf("CreateDocument: %v", err)
	}

	ref := core.NewExternalRef("notion", "page-abc")
	s := &core.Section{
		ID:          core.NewID().String(),
		DocID:       docID,
		Ref:         ref,
		Type:        core.SectionHeading,
		Title:       "External Section",
		Content:     "",
		ContentHash: "ext-hash-123",
		ContentType: core.ContentExternal,
		Metadata:    `{"system":"notion","page_id":"abc"}`,
		Order:       1,
	}
	if err := repo.CreateSection(ctx, s); err != nil {
		t.Fatalf("CreateSection external: %v", err)
	}

	sections, err := repo.ListSections(ctx, docID)
	if err != nil {
		t.Fatalf("ListSections: %v", err)
	}
	if len(sections) != 1 {
		t.Fatalf("expected 1 section, got %d", len(sections))
	}
	if sections[0].ContentType != core.ContentExternal {
		t.Errorf("ContentType = %q, want %q", sections[0].ContentType, core.ContentExternal)
	}
	if sections[0].Metadata != `{"system":"notion","page_id":"abc"}` {
		t.Errorf("Metadata = %q, want %q", sections[0].Metadata, `{"system":"notion","page_id":"abc"}`)
	}
	if sections[0].ContentHash != "ext-hash-123" {
		t.Errorf("ContentHash = %q, want %q", sections[0].ContentHash, "ext-hash-123")
	}
}

func TestListSections_ReturnsContentTypeAndMetadata(t *testing.T) {
	t.Parallel()
	db := setupDocTestDB(t)
	repo := store.NewDocumentRepo(db)
	ctx := context.Background()

	docID := core.NewID().String()
	doc := &core.Document{ID: docID, Title: "Doc", OwnerID: "user-1", Status: "active", Source: "native"}
	if err := repo.CreateDocument(ctx, doc); err != nil {
		t.Fatalf("CreateDocument: %v", err)
	}

	// Native section
	nativeRef, _ := core.ParseRef("@a1")
	native := &core.Section{
		ID:          core.NewID().String(),
		DocID:       docID,
		Ref:         nativeRef,
		Type:        core.SectionHeading,
		Title:       "Native",
		Content:     "native content",
		ContentHash: core.ContentHash("native content"),
		ContentType: core.ContentNative,
		Metadata:    "{}",
		Order:       1,
	}
	if err := repo.CreateSection(ctx, native); err != nil {
		t.Fatalf("CreateSection native: %v", err)
	}

	// External section
	extRef := core.NewExternalRef("jira", "PROJ-42")
	external := &core.Section{
		ID:          core.NewID().String(),
		DocID:       docID,
		Ref:         extRef,
		Type:        core.SectionHeading,
		Title:       "External",
		Content:     "",
		ContentHash: "jira-hash-42",
		ContentType: core.ContentExternal,
		Metadata:    `{"system":"jira","issue_key":"PROJ-42"}`,
		Order:       2,
	}
	if err := repo.CreateSection(ctx, external); err != nil {
		t.Fatalf("CreateSection external: %v", err)
	}

	sections, err := repo.ListSections(ctx, docID)
	if err != nil {
		t.Fatalf("ListSections: %v", err)
	}
	if len(sections) != 2 {
		t.Fatalf("expected 2 sections, got %d", len(sections))
	}

	// First section: native
	if sections[0].ContentType != core.ContentNative {
		t.Errorf("sections[0].ContentType = %q, want %q", sections[0].ContentType, core.ContentNative)
	}
	if sections[0].Metadata != "{}" {
		t.Errorf("sections[0].Metadata = %q, want %q", sections[0].Metadata, "{}")
	}

	// Second section: external
	if sections[1].ContentType != core.ContentExternal {
		t.Errorf("sections[1].ContentType = %q, want %q", sections[1].ContentType, core.ContentExternal)
	}
	if sections[1].Metadata != `{"system":"jira","issue_key":"PROJ-42"}` {
		t.Errorf("sections[1].Metadata = %q, want %q", sections[1].Metadata, `{"system":"jira","issue_key":"PROJ-42"}`)
	}
}

func TestFindSectionByRef_External(t *testing.T) {
	t.Parallel()
	db := setupDocTestDB(t)
	repo := store.NewDocumentRepo(db)
	ctx := context.Background()

	docID := core.NewID().String()
	doc := &core.Document{ID: docID, Title: "Doc", OwnerID: "user-1", Status: "active", Source: "native"}
	if err := repo.CreateDocument(ctx, doc); err != nil {
		t.Fatalf("CreateDocument: %v", err)
	}

	extRef := core.NewExternalRef("test", "t1")
	sectionID := core.NewID().String()
	s := &core.Section{
		ID:          sectionID,
		DocID:       docID,
		Ref:         extRef,
		Type:        core.SectionHeading,
		Title:       "External Find",
		Content:     "",
		ContentHash: "ext-find-hash",
		ContentType: core.ContentExternal,
		Metadata:    `{"system":"test","id":"t1"}`,
		Order:       1,
	}
	if err := repo.CreateSection(ctx, s); err != nil {
		t.Fatalf("CreateSection: %v", err)
	}

	got, err := repo.FindSectionByRef(ctx, docID, "@ext:test/t1")
	if err != nil {
		t.Fatalf("FindSectionByRef: %v", err)
	}
	if got.ID != sectionID {
		t.Errorf("ID = %q, want %q", got.ID, sectionID)
	}
	if got.ContentType != core.ContentExternal {
		t.Errorf("ContentType = %q, want %q", got.ContentType, core.ContentExternal)
	}
	if got.Metadata != `{"system":"test","id":"t1"}` {
		t.Errorf("Metadata = %q, want %q", got.Metadata, `{"system":"test","id":"t1"}`)
	}
	if got.ContentHash != "ext-find-hash" {
		t.Errorf("ContentHash = %q, want %q", got.ContentHash, "ext-find-hash")
	}
}

func TestFindSectionByRefGlobal_External(t *testing.T) {
	t.Parallel()
	db := setupDocTestDB(t)
	repo := store.NewDocumentRepo(db)
	ctx := context.Background()

	docID := core.NewID().String()
	doc := &core.Document{ID: docID, Title: "Doc", OwnerID: "user-1", Status: "active", Source: "native"}
	if err := repo.CreateDocument(ctx, doc); err != nil {
		t.Fatalf("CreateDocument: %v", err)
	}

	extRef := core.NewExternalRef("test", "t1")
	sectionID := core.NewID().String()
	s := &core.Section{
		ID:          sectionID,
		DocID:       docID,
		Ref:         extRef,
		Type:        core.SectionHeading,
		Title:       "External Global Find",
		Content:     "",
		ContentHash: "ext-global-hash",
		ContentType: core.ContentExternal,
		Metadata:    `{"system":"test","id":"t1"}`,
		Order:       1,
	}
	if err := repo.CreateSection(ctx, s); err != nil {
		t.Fatalf("CreateSection: %v", err)
	}

	got, gotDocID, err := repo.FindSectionByRefGlobal(ctx, "@ext:test/t1")
	if err != nil {
		t.Fatalf("FindSectionByRefGlobal: %v", err)
	}
	if got.ID != sectionID {
		t.Errorf("ID = %q, want %q", got.ID, sectionID)
	}
	if gotDocID != docID {
		t.Errorf("DocID = %q, want %q", gotDocID, docID)
	}
	if got.ContentType != core.ContentExternal {
		t.Errorf("ContentType = %q, want %q", got.ContentType, core.ContentExternal)
	}
	if got.Metadata != `{"system":"test","id":"t1"}` {
		t.Errorf("Metadata = %q, want %q", got.Metadata, `{"system":"test","id":"t1"}`)
	}
	if got.ContentHash != "ext-global-hash" {
		t.Errorf("ContentHash = %q, want %q", got.ContentHash, "ext-global-hash")
	}
}

func TestUpdateSectionContent_ExternalHashPush(t *testing.T) {
	t.Parallel()
	db := setupDocTestDB(t)
	repo := store.NewDocumentRepo(db)
	ctx := context.Background()

	docID := core.NewID().String()
	doc := &core.Document{ID: docID, Title: "Doc", OwnerID: "user-1", Status: "active", Source: "native"}
	if err := repo.CreateDocument(ctx, doc); err != nil {
		t.Fatalf("CreateDocument: %v", err)
	}

	extRef := core.NewExternalRef("github", "pr-99")
	sectionID := core.NewID().String()
	s := &core.Section{
		ID:          sectionID,
		DocID:       docID,
		Ref:         extRef,
		Type:        core.SectionHeading,
		Title:       "External Hash Push",
		Content:     "",
		ContentHash: "old-hash",
		ContentType: core.ContentExternal,
		Metadata:    `{"system":"github","pr":"99"}`,
		Order:       1,
	}
	if err := repo.CreateSection(ctx, s); err != nil {
		t.Fatalf("CreateSection: %v", err)
	}

	// Push a new hash (content stays empty for external sections)
	if err := repo.UpdateSectionContent(ctx, sectionID, "", "new-hash"); err != nil {
		t.Fatalf("UpdateSectionContent: %v", err)
	}

	// Verify the hash was updated
	sections, err := repo.ListSections(ctx, docID)
	if err != nil {
		t.Fatalf("ListSections: %v", err)
	}
	if len(sections) != 1 {
		t.Fatalf("expected 1 section, got %d", len(sections))
	}
	if sections[0].ContentHash != "new-hash" {
		t.Errorf("ContentHash = %q, want %q", sections[0].ContentHash, "new-hash")
	}

	// Verify a version was created
	versions, err := repo.GetSectionVersions(ctx, sectionID)
	if err != nil {
		t.Fatalf("GetSectionVersions: %v", err)
	}
	if len(versions) != 1 {
		t.Fatalf("expected 1 version, got %d", len(versions))
	}
	if versions[0].ContentHash != "new-hash" {
		t.Errorf("version ContentHash = %q, want %q", versions[0].ContentHash, "new-hash")
	}
}

// --- Global ref counter tests ---

func TestNextRefSeq_FirstCall(t *testing.T) {
	t.Parallel()
	db := setupDocTestDB(t)
	repo := store.NewDocumentRepo(db)
	ctx := context.Background()

	// Fresh DB with no sections — seed should be 1
	first, err := repo.NextRefSeq(ctx, 1)
	if err != nil {
		t.Fatalf("NextRefSeq: %v", err)
	}
	if first != 1 {
		t.Errorf("first seq = %d, want 1", first)
	}
}

func TestNextRefSeq_ReservesRange(t *testing.T) {
	t.Parallel()
	db := setupDocTestDB(t)
	repo := store.NewDocumentRepo(db)
	ctx := context.Background()

	// Reserve 3 — should return 1 (start of range)
	first, err := repo.NextRefSeq(ctx, 3)
	if err != nil {
		t.Fatalf("NextRefSeq(3): %v", err)
	}
	if first != 1 {
		t.Errorf("first seq = %d, want 1", first)
	}

	// Next call should return 4 (1+3)
	second, err := repo.NextRefSeq(ctx, 1)
	if err != nil {
		t.Fatalf("NextRefSeq(1): %v", err)
	}
	if second != 4 {
		t.Errorf("second seq = %d, want 4", second)
	}
}

func TestNextRefSeq_AfterExistingSections(t *testing.T) {
	t.Parallel()
	db := setupDocTestDB(t)
	repo := store.NewDocumentRepo(db)
	ctx := context.Background()

	docID := core.NewID().String()
	doc := &core.Document{ID: docID, Title: "Doc", OwnerID: "user-1", Status: "active", Source: "native"}
	if err := repo.CreateDocument(ctx, doc); err != nil {
		t.Fatalf("CreateDocument: %v", err)
	}

	// Insert 5 sections across two documents to seed the counter
	doc2ID := core.NewID().String()
	doc2 := &core.Document{ID: doc2ID, Title: "Doc2", OwnerID: "user-1", Status: "active", Source: "native"}
	if err := repo.CreateDocument(ctx, doc2); err != nil {
		t.Fatalf("CreateDocument doc2: %v", err)
	}

	refs := []string{"@a1", "@b2", "@c3"}
	for i, r := range refs {
		ref, _ := core.ParseRef(r)
		s := &core.Section{
			ID: core.NewID().String(), DocID: docID, Ref: ref,
			Type: core.SectionHeading, Content: r, ContentHash: core.ContentHash(r), Order: i,
		}
		if err := repo.CreateSection(ctx, s); err != nil {
			t.Fatalf("CreateSection %s: %v", r, err)
		}
	}
	refs2 := []string{"@d4", "@e5"}
	for i, r := range refs2 {
		ref, _ := core.ParseRef(r)
		s := &core.Section{
			ID: core.NewID().String(), DocID: doc2ID, Ref: ref,
			Type: core.SectionHeading, Content: r, ContentHash: core.ContentHash(r), Order: i,
		}
		if err := repo.CreateSection(ctx, s); err != nil {
			t.Fatalf("CreateSection %s: %v", r, err)
		}
	}

	// The migration seeds ref_counter from COUNT(*) of sections.
	// But the migration runs at setupDocTestDB time (before sections exist).
	// So the seed is 1. After inserting 5 sections, we haven't touched ref_counter yet.
	// NextRefSeq should return whatever was seeded (1), since sections were added after migration.
	first, err := repo.NextRefSeq(ctx, 1)
	if err != nil {
		t.Fatalf("NextRefSeq: %v", err)
	}
	// On fresh DB, seed is 1 (no sections at migration time), so first call returns 1
	if first != 1 {
		t.Errorf("first seq = %d, want 1", first)
	}
}

func TestDeleteDocument(t *testing.T) {
	t.Parallel()
	db := setupDocTestDB(t)
	repo := store.NewDocumentRepo(db)
	ctx := context.Background()

	docID := core.NewID().String()
	doc := &core.Document{ID: docID, Title: "To Delete", OwnerID: "user-1", Status: core.DocumentActive, Source: "native"}
	if err := repo.CreateDocument(ctx, doc); err != nil {
		t.Fatalf("CreateDocument: %v", err)
	}

	// Add sections with tags and versions
	ref, _ := core.ParseRef("@a1")
	sectionID := core.NewID().String()
	s := &core.Section{
		ID: sectionID, DocID: docID, Ref: ref, Type: core.SectionHeading,
		Content: "content", ContentHash: core.ContentHash("content"), Order: 1,
	}
	if err := repo.CreateSection(ctx, s); err != nil {
		t.Fatalf("CreateSection: %v", err)
	}
	if err := repo.AddTag(ctx, sectionID, "api"); err != nil {
		t.Fatalf("AddTag: %v", err)
	}
	if err := repo.UpdateSectionContent(ctx, sectionID, "v2", core.ContentHash("v2")); err != nil {
		t.Fatalf("UpdateSectionContent: %v", err)
	}

	// Delete the document
	if err := repo.DeleteDocument(ctx, docID); err != nil {
		t.Fatalf("DeleteDocument: %v", err)
	}

	// Document should be gone
	_, err := repo.FindDocumentByID(ctx, docID)
	if !errors.Is(err, core.ErrNotFound{}) {
		t.Errorf("expected ErrNotFound after delete, got: %v", err)
	}

	// Sections should be gone
	sections, err := repo.ListSections(ctx, docID)
	if err != nil {
		t.Fatalf("ListSections: %v", err)
	}
	if len(sections) != 0 {
		t.Errorf("expected 0 sections after delete, got %d", len(sections))
	}
}

func TestDeleteDocument_NotFound(t *testing.T) {
	t.Parallel()
	db := setupDocTestDB(t)
	repo := store.NewDocumentRepo(db)
	ctx := context.Background()

	err := repo.DeleteDocument(ctx, "does-not-exist")
	if !errors.Is(err, core.ErrNotFound{}) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestArchiveDocument(t *testing.T) {
	t.Parallel()
	db := setupDocTestDB(t)
	repo := store.NewDocumentRepo(db)
	ctx := context.Background()

	docID := core.NewID().String()
	doc := &core.Document{ID: docID, Title: "To Archive", OwnerID: "user-1", Status: core.DocumentActive, Source: "native"}
	if err := repo.CreateDocument(ctx, doc); err != nil {
		t.Fatalf("CreateDocument: %v", err)
	}

	if err := repo.ArchiveDocument(ctx, docID); err != nil {
		t.Fatalf("ArchiveDocument: %v", err)
	}

	got, err := repo.FindDocumentByID(ctx, docID)
	if err != nil {
		t.Fatalf("FindDocumentByID: %v", err)
	}
	if got.Status != core.DocumentArchived {
		t.Errorf("Status = %q, want %q", got.Status, core.DocumentArchived)
	}
}

func TestArchiveDocument_NotFound(t *testing.T) {
	t.Parallel()
	db := setupDocTestDB(t)
	repo := store.NewDocumentRepo(db)
	ctx := context.Background()

	err := repo.ArchiveDocument(ctx, "does-not-exist")
	if !errors.Is(err, core.ErrNotFound{}) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestCreateSection_ExternalDefaultMetadata(t *testing.T) {
	t.Parallel()
	db := setupDocTestDB(t)
	repo := store.NewDocumentRepo(db)
	ctx := context.Background()

	docID := core.NewID().String()
	doc := &core.Document{ID: docID, Title: "Doc", OwnerID: "user-1", Status: "active", Source: "native"}
	if err := repo.CreateDocument(ctx, doc); err != nil {
		t.Fatalf("CreateDocument: %v", err)
	}

	extRef := core.NewExternalRef("slack", "msg-1")
	s := &core.Section{
		ID:          core.NewID().String(),
		DocID:       docID,
		Ref:         extRef,
		Type:        core.SectionHeading,
		Title:       "External Default Meta",
		Content:     "",
		ContentHash: "slack-hash",
		ContentType: core.ContentExternal,
		Metadata:    "", // empty — should default to "{}"
		Order:       1,
	}
	if err := repo.CreateSection(ctx, s); err != nil {
		t.Fatalf("CreateSection: %v", err)
	}

	sections, err := repo.ListSections(ctx, docID)
	if err != nil {
		t.Fatalf("ListSections: %v", err)
	}
	if len(sections) != 1 {
		t.Fatalf("expected 1 section, got %d", len(sections))
	}
	if sections[0].Metadata != "{}" {
		t.Errorf("Metadata = %q, want %q", sections[0].Metadata, "{}")
	}
}
