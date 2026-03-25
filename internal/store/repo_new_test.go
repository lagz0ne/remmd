package store_test

import (
	"context"
	"testing"

	"github.com/lagz0ne/remmd/internal/core"
	"github.com/lagz0ne/remmd/internal/store"
)

func TestCreateSections_Batch(t *testing.T) {
	t.Parallel()
	db := setupDocTestDB(t)
	repo := store.NewDocumentRepo(db)
	ctx := context.Background()

	docID := core.NewID().String()
	doc := &core.Document{ID: docID, Title: "Batch Doc", OwnerID: "user-1", Status: core.DocumentActive, Source: "native"}
	if err := repo.CreateDocument(ctx, doc); err != nil {
		t.Fatalf("CreateDocument: %v", err)
	}

	// Reserve ref sequences
	startSeq, err := repo.NextRefSeq(ctx, 5)
	if err != nil {
		t.Fatalf("NextRefSeq: %v", err)
	}

	sections := make([]core.Section, 5)
	for i := 0; i < 5; i++ {
		ref := core.NewRef(docID, startSeq+i)
		sections[i] = core.Section{
			ID:          core.NewID().String(),
			DocID:       docID,
			Ref:         ref,
			Type:        core.SectionHeading,
			Kind:        "requirement",
			Title:       "Section " + ref.String(),
			Content:     "Content for " + ref.String(),
			ContentHash: core.ContentHash("Content for " + ref.String()),
			Order:       i,
		}
	}

	if err := repo.CreateSections(ctx, sections); err != nil {
		t.Fatalf("CreateSections: %v", err)
	}

	// Verify all 5 persisted
	got, err := repo.ListSections(ctx, docID)
	if err != nil {
		t.Fatalf("ListSections: %v", err)
	}
	if len(got) != 5 {
		t.Fatalf("expected 5 sections, got %d", len(got))
	}
	// Verify kind was persisted
	for _, s := range got {
		if s.Kind != "requirement" {
			t.Errorf("section %s kind = %q, want %q", s.ID, s.Kind, "requirement")
		}
	}
}

func TestListDocumentsWithSectionCounts(t *testing.T) {
	t.Parallel()
	db := setupDocTestDB(t)
	repo := store.NewDocumentRepo(db)
	ctx := context.Background()

	// Doc 1: 3 sections
	doc1ID := core.NewID().String()
	doc1 := &core.Document{ID: doc1ID, Title: "Doc1", OwnerID: "user-1", Status: core.DocumentActive, Source: "native", DocType: "spec"}
	if err := repo.CreateDocument(ctx, doc1); err != nil {
		t.Fatalf("CreateDocument doc1: %v", err)
	}
	sections1 := core.Parse(doc1ID, "# A\n# B\n# C", 0)
	for i := range sections1 {
		if err := repo.CreateSection(ctx, &sections1[i]); err != nil {
			t.Fatalf("CreateSection doc1[%d]: %v", i, err)
		}
	}

	// Doc 2: 1 section
	doc2ID := core.NewID().String()
	doc2 := &core.Document{ID: doc2ID, Title: "Doc2", OwnerID: "user-1", Status: core.DocumentActive, Source: "native"}
	if err := repo.CreateDocument(ctx, doc2); err != nil {
		t.Fatalf("CreateDocument doc2: %v", err)
	}
	sections2 := core.Parse(doc2ID, "# Only One", 0)
	for i := range sections2 {
		if err := repo.CreateSection(ctx, &sections2[i]); err != nil {
			t.Fatalf("CreateSection doc2[%d]: %v", i, err)
		}
	}

	summaries, err := repo.ListDocumentsWithSectionCounts(ctx)
	if err != nil {
		t.Fatalf("ListDocumentsWithSectionCounts: %v", err)
	}
	if len(summaries) != 2 {
		t.Fatalf("expected 2 summaries, got %d", len(summaries))
	}

	// Ordered by created_at, so doc1 should come first
	counts := map[string]int{}
	for _, s := range summaries {
		counts[s.Document.ID] = s.SectionCount
	}
	if counts[doc1ID] != 3 {
		t.Errorf("doc1 section count = %d, want 3", counts[doc1ID])
	}
	if counts[doc2ID] != 1 {
		t.Errorf("doc2 section count = %d, want 1", counts[doc2ID])
	}
}

func TestSearchSections_FTS(t *testing.T) {
	t.Parallel()
	db := setupDocTestDB(t)
	repo := store.NewDocumentRepo(db)
	ctx := context.Background()

	docID := core.NewID().String()
	doc := &core.Document{ID: docID, Title: "FTS Doc", OwnerID: "user-1", Status: core.DocumentActive, Source: "native"}
	if err := repo.CreateDocument(ctx, doc); err != nil {
		t.Fatalf("CreateDocument: %v", err)
	}

	ref1, _ := core.ParseRef("@a1")
	ref2, _ := core.ParseRef("@b2")
	ref3, _ := core.ParseRef("@c3")
	s1 := &core.Section{ID: core.NewID().String(), DocID: docID, Ref: ref1, Type: core.SectionHeading, Title: "Authentication", Content: "OAuth2 authentication flow", ContentHash: core.ContentHash("OAuth2 authentication flow"), Order: 0}
	s2 := &core.Section{ID: core.NewID().String(), DocID: docID, Ref: ref2, Type: core.SectionHeading, Title: "Authorization", Content: "RBAC authorization rules", ContentHash: core.ContentHash("RBAC authorization rules"), Order: 1}
	s3 := &core.Section{ID: core.NewID().String(), DocID: docID, Ref: ref3, Type: core.SectionHeading, Title: "Database", Content: "PostgreSQL database schema", ContentHash: core.ContentHash("PostgreSQL database schema"), Order: 2}

	for _, s := range []*core.Section{s1, s2, s3} {
		if err := repo.CreateSection(ctx, s); err != nil {
			t.Fatalf("CreateSection %s: %v", s.ID, err)
		}
	}

	// Search for "authentication"
	results, err := repo.SearchSections(ctx, "authentication")
	if err != nil {
		t.Fatalf("SearchSections(authentication): %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result for 'authentication', got %d", len(results))
	}
	if results[0].ID != s1.ID {
		t.Errorf("expected section %s, got %s", s1.ID, results[0].ID)
	}

	// Search for "database"
	results, err = repo.SearchSections(ctx, "database")
	if err != nil {
		t.Fatalf("SearchSections(database): %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result for 'database', got %d", len(results))
	}
	if results[0].ID != s3.ID {
		t.Errorf("expected section %s, got %s", s3.ID, results[0].ID)
	}
}

func TestLinksContainingSections_Batch(t *testing.T) {
	t.Parallel()
	db := setupLinkTestDB(t)
	linkRepo := store.NewLinkRepo(db)
	ctx := context.Background()

	// Create three sections in different docs
	secA := createTestSection(t, db, "Doc A "+core.NewID().String(), "@a1", "# Section A")
	secB := createTestSection(t, db, "Doc B "+core.NewID().String(), "@a1", "# Section B")
	secC := createTestSection(t, db, "Doc C "+core.NewID().String(), "@a1", "# Section C")

	// link1: secA -> secB
	link1 := core.NewLink(
		[]string{secA.ID}, []string{secB.ID},
		core.RelImplements,
		core.Rationale{Claim: "impl", Scope: "all", Exclusions: ""},
		"user-1",
	)
	if err := linkRepo.CreateLink(ctx, link1); err != nil {
		t.Fatalf("CreateLink link1: %v", err)
	}

	// link2: secA -> secC
	link2 := core.NewLink(
		[]string{secA.ID}, []string{secC.ID},
		core.RelTests,
		core.Rationale{Claim: "tests", Scope: "unit", Exclusions: ""},
		"user-2",
	)
	if err := linkRepo.CreateLink(ctx, link2); err != nil {
		t.Fatalf("CreateLink link2: %v", err)
	}

	// Batch query for all three
	result, err := linkRepo.LinksContainingSections(ctx, []string{secA.ID, secB.ID, secC.ID})
	if err != nil {
		t.Fatalf("LinksContainingSections: %v", err)
	}

	// secA should appear in 2 links
	if len(result[secA.ID]) != 2 {
		t.Errorf("secA links = %d, want 2", len(result[secA.ID]))
	}
	// secB should appear in 1 link
	if len(result[secB.ID]) != 1 {
		t.Errorf("secB links = %d, want 1", len(result[secB.ID]))
	}
	// secC should appear in 1 link
	if len(result[secC.ID]) != 1 {
		t.Errorf("secC links = %d, want 1", len(result[secC.ID]))
	}

	// Empty input returns nil
	empty, err := linkRepo.LinksContainingSections(ctx, nil)
	if err != nil {
		t.Fatalf("LinksContainingSections(nil): %v", err)
	}
	if empty != nil {
		t.Errorf("expected nil for empty input, got %v", empty)
	}
}

func TestRelationRepo_CRUD(t *testing.T) {
	t.Parallel()
	db := setupDocTestDB(t)
	docRepo := store.NewDocumentRepo(db)
	relRepo := store.NewRelationRepo(db)
	ctx := context.Background()

	// Create two docs
	doc1 := &core.Document{ID: core.NewID().String(), Title: "Doc1", OwnerID: "u1", Status: core.DocumentActive, Source: "native"}
	doc2 := &core.Document{ID: core.NewID().String(), Title: "Doc2", OwnerID: "u1", Status: core.DocumentActive, Source: "native"}
	for _, d := range []*core.Document{doc1, doc2} {
		if err := docRepo.CreateDocument(ctx, d); err != nil {
			t.Fatalf("CreateDocument: %v", err)
		}
	}

	// Create relation
	rel := core.NewRelation(doc1.ID, doc2.ID, "depends-on")
	if err := relRepo.CreateRelation(ctx, rel); err != nil {
		t.Fatalf("CreateRelation: %v", err)
	}

	// ListRelationsFrom
	from, err := relRepo.ListRelationsFrom(ctx, doc1.ID)
	if err != nil {
		t.Fatalf("ListRelationsFrom: %v", err)
	}
	if len(from) != 1 {
		t.Fatalf("expected 1 relation from doc1, got %d", len(from))
	}
	if from[0].ID != rel.ID {
		t.Errorf("relation ID = %q, want %q", from[0].ID, rel.ID)
	}
	if from[0].RelationType != "depends-on" {
		t.Errorf("relation type = %q, want %q", from[0].RelationType, "depends-on")
	}

	// ListRelationsTo
	to, err := relRepo.ListRelationsTo(ctx, doc2.ID)
	if err != nil {
		t.Fatalf("ListRelationsTo: %v", err)
	}
	if len(to) != 1 {
		t.Fatalf("expected 1 relation to doc2, got %d", len(to))
	}

	// Delete
	if err := relRepo.DeleteRelation(ctx, rel.ID); err != nil {
		t.Fatalf("DeleteRelation: %v", err)
	}

	// After delete, should be empty
	from, err = relRepo.ListRelationsFrom(ctx, doc1.ID)
	if err != nil {
		t.Fatalf("ListRelationsFrom after delete: %v", err)
	}
	if len(from) != 0 {
		t.Errorf("expected 0 relations after delete, got %d", len(from))
	}
}

func TestTemplateRepo_CRUD(t *testing.T) {
	t.Parallel()
	db := setupDocTestDB(t)
	repo := store.NewTemplateRepo(db)
	ctx := context.Background()

	// Set two templates for "spec"
	t1 := core.SchemaTemplate{DocType: "spec", RequiredKind: "goal", MinCount: 1}
	t2 := core.SchemaTemplate{DocType: "spec", RequiredKind: "requirement", MinCount: 3}
	for _, tmpl := range []core.SchemaTemplate{t1, t2} {
		if err := repo.SetTemplate(ctx, tmpl); err != nil {
			t.Fatalf("SetTemplate %s/%s: %v", tmpl.DocType, tmpl.RequiredKind, err)
		}
	}

	// Get templates
	templates, err := repo.GetTemplates(ctx, "spec")
	if err != nil {
		t.Fatalf("GetTemplates: %v", err)
	}
	if len(templates) != 2 {
		t.Fatalf("expected 2 templates, got %d", len(templates))
	}

	// Upsert: update min_count
	t1Updated := core.SchemaTemplate{DocType: "spec", RequiredKind: "goal", MinCount: 2}
	if err := repo.SetTemplate(ctx, t1Updated); err != nil {
		t.Fatalf("SetTemplate (upsert): %v", err)
	}
	templates, err = repo.GetTemplates(ctx, "spec")
	if err != nil {
		t.Fatalf("GetTemplates after upsert: %v", err)
	}
	for _, tmpl := range templates {
		if tmpl.RequiredKind == "goal" && tmpl.MinCount != 2 {
			t.Errorf("goal min_count = %d, want 2 after upsert", tmpl.MinCount)
		}
	}

	// Delete one
	if err := repo.DeleteTemplate(ctx, "spec", "goal"); err != nil {
		t.Fatalf("DeleteTemplate: %v", err)
	}
	templates, err = repo.GetTemplates(ctx, "spec")
	if err != nil {
		t.Fatalf("GetTemplates after delete: %v", err)
	}
	if len(templates) != 1 {
		t.Fatalf("expected 1 template after delete, got %d", len(templates))
	}
	if templates[0].RequiredKind != "requirement" {
		t.Errorf("remaining template kind = %q, want %q", templates[0].RequiredKind, "requirement")
	}

	// Different doc type should be empty
	other, err := repo.GetTemplates(ctx, "runbook")
	if err != nil {
		t.Fatalf("GetTemplates(runbook): %v", err)
	}
	if len(other) != 0 {
		t.Errorf("expected 0 templates for runbook, got %d", len(other))
	}
}

func TestSnapshotService_ComputeSnapshot(t *testing.T) {
	t.Parallel()
	db := setupDocTestDB(t)
	docRepo := store.NewDocumentRepo(db)
	linkRepo := store.NewLinkRepo(db)
	ctx := context.Background()

	// Create two docs with sections
	secA := createTestSection(t, db, "Doc A "+core.NewID().String(), "@a1", "# Left Side")
	secB := createTestSection(t, db, "Doc B "+core.NewID().String(), "@a1", "# Right Side")

	// Create a link
	link := core.NewLink(
		[]string{secA.ID}, []string{secB.ID},
		core.RelImplements,
		core.Rationale{Claim: "impl", Scope: "all", Exclusions: ""},
		"user-1",
	)
	if err := linkRepo.CreateLink(ctx, link); err != nil {
		t.Fatalf("CreateLink: %v", err)
	}

	// Compute snapshot
	svc := store.NewSnapshotService(linkRepo, docRepo)
	snap, err := svc.ComputeSnapshot(ctx, link.ID)
	if err != nil {
		t.Fatalf("ComputeSnapshot: %v", err)
	}

	if snap.LinkID != link.ID {
		t.Errorf("LinkID = %q, want %q", snap.LinkID, link.ID)
	}
	if len(snap.LeftContentHashes) != 1 {
		t.Fatalf("expected 1 left hash, got %d", len(snap.LeftContentHashes))
	}
	if len(snap.RightContentHashes) != 1 {
		t.Fatalf("expected 1 right hash, got %d", len(snap.RightContentHashes))
	}

	// Verify the hashes match the sections' content hashes
	if snap.LeftContentHashes[0] != secA.ContentHash {
		t.Errorf("left hash = %q, want %q", snap.LeftContentHashes[0], secA.ContentHash)
	}
	if snap.RightContentHashes[0] != secB.ContentHash {
		t.Errorf("right hash = %q, want %q", snap.RightContentHashes[0], secB.ContentHash)
	}

	// The hash should be deterministic
	hash1 := snap.Hash()
	snap2, _ := svc.ComputeSnapshot(ctx, link.ID)
	hash2 := snap2.Hash()
	if hash1 != hash2 {
		t.Errorf("snapshot hash not deterministic: %q vs %q", hash1, hash2)
	}

	// After updating content, the hash should change
	newContent := "# Updated Left"
	newHash := core.ContentHash(newContent)
	if err := docRepo.UpdateSectionContent(ctx, secA.ID, newContent, newHash); err != nil {
		t.Fatalf("UpdateSectionContent: %v", err)
	}

	snap3, err := svc.ComputeSnapshot(ctx, link.ID)
	if err != nil {
		t.Fatalf("ComputeSnapshot after update: %v", err)
	}
	hash3 := snap3.Hash()
	if hash1 == hash3 {
		t.Error("snapshot hash should change after content update")
	}
}
