package serve

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/lagz0ne/remmd/internal/app"
	"github.com/lagz0ne/remmd/internal/core"
	"github.com/lagz0ne/remmd/internal/playbook"
	"github.com/lagz0ne/remmd/internal/store"
	"github.com/nats-io/nats.go"
)

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
			Kind        string  `json:"kind,omitempty"`
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
				Kind:        s.Kind,
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

	// remmd.q.graph — full graph: documents as nodes, links + relations as edges
	nc.Subscribe("remmd.q.graph", func(msg *nats.Msg) {
		ctx := context.Background()
		docs, err := application.Docs.ListDocuments(ctx)
		if err != nil {
			slog.Error("handler: graph list documents", "error", err)
			replyErr(msg, err)
			return
		}
		links, err := application.Links.ListLinks(ctx, "")
		if err != nil {
			slog.Error("handler: graph list links", "error", err)
			replyErr(msg, err)
			return
		}
		relations, err := application.Relations.ListAllRelations(ctx)
		if err != nil {
			slog.Error("handler: graph list relations", "error", err)
			replyErr(msg, err)
			return
		}
		resolver := func(ctx context.Context, sectionID string) (string, error) {
			sec, err := application.Docs.FindSectionByID(ctx, sectionID)
			if err != nil {
				return "", err
			}
			return sec.DocID, nil
		}
		sectionLister := func(ctx context.Context, docID string) ([]*core.Section, error) {
			return application.Docs.ListSections(ctx, docID)
		}
		reply(msg, buildGraphResponse(ctx, docs, links, relations, resolver, sectionLister))
	})

	// remmd.q.playbook — active playbook schema
	nc.Subscribe("remmd.q.playbook", func(msg *nats.Msg) {
		ctx := context.Background()
		pb, _, err := application.Playbooks.Latest(ctx, "default")
		if err != nil {
			slog.Error("handler: playbook latest", "error", err)
			replyErr(msg, err)
			return
		}
		if pb == nil {
			reply(msg, playbookResponse{})
			return
		}
		reply(msg, buildPlaybookResponse(pb))
	})

	// remmd.q.validate — run playbook validation against all documents
	nc.Subscribe("remmd.q.validate", func(msg *nats.Msg) {
		ctx := context.Background()
		pb, _, err := application.Playbooks.Latest(ctx, "default")
		if err != nil {
			slog.Error("handler: validate playbook latest", "error", err)
			replyErr(msg, err)
			return
		}
		if pb == nil {
			reply(msg, validationResponse{Diagnostics: []validationDiag{}})
			return
		}

		docs, err := application.Docs.ListDocuments(ctx)
		if err != nil {
			slog.Error("handler: validate list documents", "error", err)
			replyErr(msg, err)
			return
		}

		var nodes []playbook.Node
		for _, d := range docs {
			nodes = append(nodes, playbook.Node{
				Type: d.DocType,
				ID:   d.ID,
				Data: map[string]any{
					"_node_id": d.ID,
					"title":    d.Title,
					"status":   string(d.Status),
					"source":   d.Source,
				},
			})
		}

		relations, err := application.Relations.ListAllRelations(ctx)
		if err != nil {
			slog.Error("handler: validate list relations", "error", err)
			replyErr(msg, err)
			return
		}

		gc := newRelationGraph(relations, nodes)
		diags := playbook.RunWithGraph(pb, nodes, gc)
		reply(msg, buildValidationResponse(diags))
	})

	// remmd.q.positions — load saved node positions
	nc.Subscribe("remmd.q.positions", func(msg *nats.Msg) {
		ctx := context.Background()
		positions, err := application.Positions.LoadPositions(ctx)
		if err != nil {
			slog.Error("handler: load positions", "error", err)
			replyErr(msg, err)
			return
		}
		reply(msg, positions)
	})

	// remmd.c.positions — save node positions
	nc.Subscribe("remmd.c.positions", func(msg *nats.Msg) {
		ctx := context.Background()
		var positions []store.NodePosition
		if err := json.Unmarshal(msg.Data, &positions); err != nil {
			replyErr(msg, err)
			return
		}
		if err := application.Positions.SavePositions(ctx, positions); err != nil {
			slog.Error("handler: save positions", "error", err)
			replyErr(msg, err)
			return
		}
		reply(msg, map[string]bool{"ok": true})
	})

	// remmd.c.positions.clear — clear all saved positions
	nc.Subscribe("remmd.c.positions.clear", func(msg *nats.Msg) {
		ctx := context.Background()
		if err := application.Positions.ClearPositions(ctx); err != nil {
			slog.Error("handler: clear positions", "error", err)
			replyErr(msg, err)
			return
		}
		reply(msg, map[string]bool{"ok": true})
	})

	// remmd.q.schema — static schema
	nc.Subscribe("remmd.q.schema", func(msg *nats.Msg) {
		reply(msg, map[string]any{
			"subjects": map[string]string{
				"remmd.q.documents":            "List all documents",
				"remmd.q.documents.*.sections": "List sections for a document",
				"remmd.q.section.*":            "Get a section by ref",
				"remmd.q.graph":                "Full graph: documents as nodes, links as edges",
				"remmd.q.playbook":             "Active playbook schema (types, fields, sections, rules, edges)",
				"remmd.q.validate":             "Run playbook validation against all documents",
				"remmd.q.positions":            "Load saved node positions",
				"remmd.c.positions":            "Save node positions",
				"remmd.c.positions.clear":      "Clear all saved positions",
				"remmd.q.schema":               "This schema",
				"remmd.doc.*.section.*":        "Event: section changed (docID, ref)",
			},
		})
	})
}

var statePriority = map[string]int{
	string(core.LinkAligned): 1,
	string(core.LinkPending): 2,
	string(core.LinkStale):   3,
	string(core.LinkBroken):  4,
}

func worstState(states []string) string {
	best := 0
	result := ""
	for _, s := range states {
		if p := statePriority[s]; p > best {
			best = p
			result = s
		}
	}
	return result
}

func deriveSectionLinkState(ctx context.Context, links *store.LinkRepo, sectionID string) string {
	ls, err := links.LinksContainingSection(ctx, sectionID)
	if err != nil || len(ls) == 0 {
		return ""
	}
	return deriveLinkState(ls)
}

func deriveLinkState(links []*core.Link) string {
	states := make([]string, len(links))
	for i, l := range links {
		states[i] = string(l.State)
	}
	return worstState(states)
}

func deriveLinkInfoState(links []*core.LinkInfo) string {
	states := make([]string, len(links))
	for i, li := range links {
		states[i] = li.State
	}
	return worstState(states)
}

func subjectPart(subject string, index int) (string, error) {
	parts := strings.Split(subject, ".")
	if index >= len(parts) {
		return "", errBadSubject
	}
	return parts[index], nil
}

type sectionDocResolver func(ctx context.Context, sectionID string) (string, error)

type graphResponse struct {
	Nodes []graphNode `json:"nodes"`
	Edges []graphEdge `json:"edges"`
}

type graphNode struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Status       string `json:"status"`
	Source       string `json:"source"`
	DocType      string `json:"doc_type"`
	Brief        string `json:"brief"`
	SectionCount int    `json:"section_count"`
}

type graphEdge struct {
	ID               string   `json:"id"`
	SourceDocID      string   `json:"source_doc_id"`
	TargetDocID      string   `json:"target_doc_id"`
	State            string   `json:"state"`
	RelationshipType string   `json:"relationship_type"`
	LeftSectionIDs   []string `json:"left_section_ids"`
	RightSectionIDs  []string `json:"right_section_ids"`
	IsRelation       bool     `json:"is_relation,omitempty"`
}

type sectionLister func(ctx context.Context, docID string) ([]*core.Section, error)

func buildGraphResponse(ctx context.Context, docs []*core.Document, links []*core.Link, relations []core.Relation, resolve sectionDocResolver, listSections sectionLister) graphResponse {
	nodes := make([]graphNode, 0, len(docs))
	for _, d := range docs {
		gn := graphNode{
			ID:      d.ID,
			Title:   d.Title,
			Status:  string(d.Status),
			Source:  d.Source,
			DocType: d.DocType,
		}
		if sections, err := listSections(ctx, d.ID); err == nil {
			gn.SectionCount = len(sections)
			gn.Brief = findGoalBrief(sections)
		}
		nodes = append(nodes, gn)
	}

	edges := make([]graphEdge, 0, len(links))
	for _, l := range links {
		srcDocID, err := resolveDocID(ctx, l.LeftSectionIDs, resolve)
		if err != nil {
			continue
		}
		tgtDocID, err := resolveDocID(ctx, l.RightSectionIDs, resolve)
		if err != nil {
			continue
		}
		edges = append(edges, graphEdge{
			ID:               l.ID,
			SourceDocID:      srcDocID,
			TargetDocID:      tgtDocID,
			State:            string(l.State),
			RelationshipType: string(l.RelationshipType),
			LeftSectionIDs:   l.LeftSectionIDs,
			RightSectionIDs:  l.RightSectionIDs,
		})
	}

	for _, rel := range relations {
		edges = append(edges, graphEdge{
			ID:               rel.ID,
			SourceDocID:      rel.FromDocID,
			TargetDocID:      rel.ToDocID,
			State:            "aligned",
			RelationshipType: rel.RelationType,
			IsRelation:       true,
		})
	}

	return graphResponse{Nodes: nodes, Edges: edges}
}

// relationGraph adapts remmd relations to playbook.GraphContext for CEL evaluation.
type relationGraph struct {
	outEdges map[string][]core.Relation // from_doc_id -> relations
	inEdges  map[string][]core.Relation // to_doc_id -> relations
	nodeMap  map[string]playbook.Node   // type:id -> node
}

func newRelationGraph(relations []core.Relation, nodes []playbook.Node) *relationGraph {
	g := &relationGraph{
		outEdges: make(map[string][]core.Relation),
		inEdges:  make(map[string][]core.Relation),
		nodeMap:  make(map[string]playbook.Node),
	}
	for _, r := range relations {
		g.outEdges[r.FromDocID] = append(g.outEdges[r.FromDocID], r)
		g.inEdges[r.ToDocID] = append(g.inEdges[r.ToDocID], r)
	}
	for _, n := range nodes {
		g.nodeMap[n.Type+":"+n.ID] = n
	}
	return g
}

func (g *relationGraph) EdgesOut(nodeID string, edgeType string) []map[string]any {
	var result []map[string]any
	for _, r := range g.outEdges[nodeID] {
		if r.RelationType == edgeType {
			result = append(result, map[string]any{
				"id":        r.ID,
				"source_id": r.FromDocID,
				"target_id": r.ToDocID,
				"type":      r.RelationType,
			})
		}
	}
	return result
}

func (g *relationGraph) EdgesIn(nodeID string, edgeType string) []map[string]any {
	var result []map[string]any
	for _, r := range g.inEdges[nodeID] {
		if r.RelationType == edgeType {
			result = append(result, map[string]any{
				"id":        r.ID,
				"source_id": r.FromDocID,
				"target_id": r.ToDocID,
				"type":      r.RelationType,
			})
		}
	}
	return result
}

func (g *relationGraph) NodeExists(nodeType string, nodeID string) bool {
	_, ok := g.nodeMap[nodeType+":"+nodeID]
	return ok
}

func findGoalBrief(sections []*core.Section) string {
	for _, s := range sections {
		if strings.EqualFold(s.Title, "goal") && s.Content != "" {
			return s.Content
		}
	}
	for _, s := range sections {
		if s.Content != "" {
			return s.Content
		}
	}
	return ""
}

func resolveDocID(ctx context.Context, sectionIDs []string, resolve sectionDocResolver) (string, error) {
	if len(sectionIDs) == 0 {
		return "", errBadSubject
	}
	return resolve(ctx, sectionIDs[0])
}

type playbookResponse struct {
	Types []playbookTypeResp   `json:"types"`
	Edges []playbookEdgeResp   `json:"edges"`
	Rules []playbookRuleResp   `json:"rules"`
}

type playbookTypeResp struct {
	Name        string             `json:"name"`
	Description string             `json:"description,omitempty"`
	Fields      []playbookFieldResp `json:"fields"`
	Sections    []playbookSectResp  `json:"sections"`
	Rules       []playbookRuleResp  `json:"rules"`
}

type playbookFieldResp struct {
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	Required bool     `json:"required"`
	Target   string   `json:"target,omitempty"`
	Targets  []string `json:"targets,omitempty"`
	Values   []string `json:"values,omitempty"`
}

type playbookSectResp struct {
	Name     string `json:"name"`
	Required bool   `json:"required"`
}

type playbookEdgeResp struct {
	Name     string `json:"name"`
	Notation string `json:"notation"`
}

type playbookRuleResp struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Severity    string `json:"severity"`
	Expr        string `json:"expr"`
}

type validationResponse struct {
	Total       int              `json:"total"`
	Diagnostics []validationDiag `json:"diagnostics"`
}

type validationDiag struct {
	Rule     string `json:"rule"`
	NodeID   string `json:"node_id"`
	NodeType string `json:"node_type"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

func buildPlaybookResponse(pb *playbook.Playbook) playbookResponse {
	var resp playbookResponse

	typeNames := make([]string, 0, len(pb.Types))
	for name := range pb.Types {
		typeNames = append(typeNames, name)
	}
	slices.Sort(typeNames)

	for _, tName := range typeNames {
		td := pb.Types[tName]
		tr := playbookTypeResp{
			Name:        td.Name,
			Description: td.Description,
		}

		fieldNames := make([]string, 0, len(td.Fields))
		for fname := range td.Fields {
			fieldNames = append(fieldNames, fname)
		}
		slices.Sort(fieldNames)
		for _, fname := range fieldNames {
			fd := td.Fields[fname]
			tr.Fields = append(tr.Fields, playbookFieldResp{
				Name:     fname,
				Type:     fd.Type,
				Required: fd.Required,
				Target:   fd.Target,
				Targets:  fd.Targets,
				Values:   fd.Values,
			})
		}

		for _, sd := range td.Sections {
			tr.Sections = append(tr.Sections, playbookSectResp{
				Name:     sd.Name,
				Required: sd.Required,
			})
		}

		ruleNames := make([]string, 0, len(td.Rules))
		for rname := range td.Rules {
			ruleNames = append(ruleNames, rname)
		}
		slices.Sort(ruleNames)
		for _, rname := range ruleNames {
			rd := td.Rules[rname]
			tr.Rules = append(tr.Rules, playbookRuleResp{
				Name:        rd.Name,
				Description: rd.Description,
				Severity:    rd.Severity,
				Expr:        rd.Expr,
			})
		}
		resp.Types = append(resp.Types, tr)
	}

	edgeNames := make([]string, 0, len(pb.Edges))
	for name := range pb.Edges {
		edgeNames = append(edgeNames, name)
	}
	slices.Sort(edgeNames)
	for _, eName := range edgeNames {
		ed := pb.Edges[eName]
		maxStr := "*"
		if ed.MaxCard >= 0 {
			maxStr = fmt.Sprintf("%d", ed.MaxCard)
		}
		from := strings.Join(ed.From, " | ")
		notation := fmt.Sprintf("%s -> %s [%d..%s]", from, ed.To, ed.MinCard, maxStr)
		resp.Edges = append(resp.Edges, playbookEdgeResp{
			Name:     ed.Name,
			Notation: notation,
		})
	}

	globalRuleNames := make([]string, 0, len(pb.Rules))
	for name := range pb.Rules {
		globalRuleNames = append(globalRuleNames, name)
	}
	slices.Sort(globalRuleNames)
	for _, rName := range globalRuleNames {
		rd := pb.Rules[rName]
		resp.Rules = append(resp.Rules, playbookRuleResp{
			Name:        rd.Name,
			Description: rd.Description,
			Severity:    rd.Severity,
			Expr:        rd.Expr,
		})
	}

	return resp
}

func buildValidationResponse(diags []playbook.Diagnostic) validationResponse {
	out := make([]validationDiag, 0, len(diags))
	for _, d := range diags {
		out = append(out, validationDiag{
			Rule:     d.Rule,
			NodeID:   d.NodeID,
			NodeType: d.NodeType,
			Severity: d.Severity,
			Message:  d.Message,
		})
	}
	return validationResponse{
		Total:       len(out),
		Diagnostics: out,
	}
}

var errBadSubject = &subjectError{"malformed subject"}

type subjectError struct{ msg string }

func (e *subjectError) Error() string { return e.msg }
