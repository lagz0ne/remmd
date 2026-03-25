package serve

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"

	"github.com/lagz0ne/remmd/internal/app"
	"github.com/lagz0ne/remmd/internal/core"
	"github.com/lagz0ne/remmd/internal/store"
	"github.com/nats-io/nats.go"
)

// registerHandlers wires NATS request-reply subjects to real app repos.
func registerHandlers(nc *nats.Conn, application *app.App) {
	reply := func(msg *nats.Msg, data any) {
		b, _ := json.Marshal(data)
		msg.Respond(b)
	}
	replyErr := func(msg *nats.Msg, err error) {
		reply(msg, map[string]string{"error": err.Error()})
	}

	// remmd.q.documents — list all documents
	nc.Subscribe("remmd.q.documents", func(msg *nats.Msg) {
		ctx := context.Background()
		docs, err := application.Docs.ListDocuments(ctx)
		if err != nil {
			slog.Error("handler: list documents", "error", err)
			replyErr(msg, err)
			return
		}

		type docResponse struct {
			ID     string `json:"id"`
			Title  string `json:"title"`
			Status string `json:"status"`
			Source string `json:"source"`
		}
		result := make([]docResponse, 0, len(docs))
		for _, d := range docs {
			result = append(result, docResponse{
				ID:     d.ID,
				Title:  d.Title,
				Status: string(d.Status),
				Source: d.Source,
			})
		}
		reply(msg, result)
	})

	// remmd.q.documents.*.sections — list sections for a document
	nc.Subscribe("remmd.q.documents.*.sections", func(msg *nats.Msg) {
		ctx := context.Background()
		docID, err := subjectPart(msg.Subject, 3)
		if err != nil {
			replyErr(msg, err)
			return
		}

		sections, err := application.Docs.ListSections(ctx, docID)
		if err != nil {
			slog.Error("handler: list sections", "doc_id", docID, "error", err)
			replyErr(msg, err)
			return
		}

		// Batch-fetch link info for all sections in one query.
		sectionIDs := make([]string, len(sections))
		for i, s := range sections {
			sectionIDs[i] = s.ID
		}
		linkMap, err := application.Links.LinksContainingSections(ctx, sectionIDs)
		if err != nil {
			slog.Error("handler: batch link lookup", "doc_id", docID, "error", err)
			replyErr(msg, err)
			return
		}

		type sectionResponse struct {
			ID          string  `json:"id"`
			Ref         string  `json:"ref"`
			Type        string  `json:"type"`
			Title       string  `json:"title"`
			Content     string  `json:"content"`
			ContentHash string  `json:"content_hash"`
			ContentType string  `json:"content_type"`
			ParentRef   *string `json:"parent_ref,omitempty"`
			Order       int     `json:"order"`
			LinkState   string  `json:"link_state,omitempty"`
		}

		result := make([]sectionResponse, 0, len(sections))
		for _, s := range sections {
			ls := deriveLinkInfoState(linkMap[s.ID])
			var parentRef *string
			if s.ParentRef != nil {
				pr := s.ParentRef.String()
				parentRef = &pr
			}
			result = append(result, sectionResponse{
				ID:          s.ID,
				Ref:         s.Ref.String(),
				Type:        string(s.Type),
				Title:       s.Title,
				Content:     s.Content,
				ContentHash: s.ContentHash,
				ContentType: string(s.ContentType),
				ParentRef:   parentRef,
				Order:       s.Order,
				LinkState:   ls,
			})
		}

		reply(msg, map[string]any{
			"doc_id":   docID,
			"sections": result,
		})
	})

	// remmd.q.section.* — single section by ref
	nc.Subscribe("remmd.q.section.*", func(msg *nats.Msg) {
		ctx := context.Background()
		// Refs like @a1 have no dots, so parts[3] is the full ref.
		// For refs with dots we'd need `>` wildcard; handle native refs for now.
		ref, err := subjectPart(msg.Subject, 3)
		if err != nil {
			replyErr(msg, err)
			return
		}

		section, docID, err := application.Docs.FindSectionByRefGlobal(ctx, ref)
		if err != nil {
			slog.Error("handler: find section by ref", "ref", ref, "error", err)
			replyErr(msg, err)
			return
		}

		ls := deriveSectionLinkState(ctx, application.Links, section.ID)

		reply(msg, map[string]any{
			"id":           section.ID,
			"doc_id":       docID,
			"ref":          section.Ref.String(),
			"type":         string(section.Type),
			"title":        section.Title,
			"content":      section.Content,
			"content_hash": section.ContentHash,
			"content_type": string(section.ContentType),
			"link_state":   ls,
		})
	})

	// remmd.q.schema — static schema
	nc.Subscribe("remmd.q.schema", func(msg *nats.Msg) {
		reply(msg, map[string]any{
			"subjects": map[string]string{
				"remmd.q.documents":            "List all documents",
				"remmd.q.documents.*.sections": "List sections for a document",
				"remmd.q.section.*":            "Get a section by ref",
				"remmd.q.schema":               "This schema",
				"remmd.doc.*.section.*":        "Event: section changed (docID, ref)",
			},
		})
	})
}

// deriveSectionLinkState queries all links containing the given section and
// returns the "worst" link state. Priority: broken > stale > pending > aligned.
// Returns empty string if the section has no links.
func deriveSectionLinkState(ctx context.Context, links *store.LinkRepo, sectionID string) string {
	ls, err := links.LinksContainingSection(ctx, sectionID)
	if err != nil || len(ls) == 0 {
		return ""
	}
	return deriveLinkState(ls)
}

// deriveLinkState returns the worst state across all provided links.
// Priority: broken > stale > pending > aligned.
// Archived links are ignored. Returns "" if no active links.
func deriveLinkState(links []*core.Link) string {
	priority := map[core.LinkState]int{
		core.LinkAligned: 1,
		core.LinkPending: 2,
		core.LinkStale:   3,
		core.LinkBroken:  4,
	}

	worst := 0
	var worstState core.LinkState
	for _, l := range links {
		if l.State == core.LinkArchived {
			continue
		}
		p := priority[l.State]
		if p > worst {
			worst = p
			worstState = l.State
		}
	}
	if worst == 0 {
		return ""
	}
	return string(worstState)
}

// subjectPart splits a NATS subject on "." and returns the part at the given index.
func subjectPart(subject string, index int) (string, error) {
	parts := strings.Split(subject, ".")
	if index >= len(parts) {
		return "", errBadSubject
	}
	return parts[index], nil
}

// deriveLinkInfoState returns the worst state across LinkInfo entries.
// Priority: broken > stale > pending > aligned. Archived links are ignored.
func deriveLinkInfoState(links []*core.LinkInfo) string {
	priority := map[string]int{
		string(core.LinkAligned): 1,
		string(core.LinkPending): 2,
		string(core.LinkStale):   3,
		string(core.LinkBroken):  4,
	}

	worst := 0
	var worstState string
	for _, li := range links {
		if li.State == string(core.LinkArchived) {
			continue
		}
		p := priority[li.State]
		if p > worst {
			worst = p
			worstState = li.State
		}
	}
	if worst == 0 {
		return ""
	}
	return worstState
}

var errBadSubject = &subjectError{"malformed subject"}

type subjectError struct{ msg string }

func (e *subjectError) Error() string { return e.msg }
