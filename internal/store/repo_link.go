package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/lagz0ne/remmd/internal/core"
)

// LinkRepo is the SQLite implementation of link persistence.
type LinkRepo struct {
	db *sql.DB
}

// NewLinkRepo creates a LinkRepo backed by the given *sql.DB.
func NewLinkRepo(db *sql.DB) *LinkRepo {
	return &LinkRepo{db: db}
}

// CreateLink persists a new link with its section IDs.
func (r *LinkRepo) CreateLink(ctx context.Context, link *core.Link) error {
	return WithTx(ctx, r.db, func(tx *sql.Tx) error {
		if link.CreatedAt.IsZero() {
			link.CreatedAt = time.Now()
		}
		_, err := tx.ExecContext(ctx,
			`INSERT INTO links (id, relationship_type, rationale_claim, rationale_scope, rationale_exclusions, state, left_intervention, right_intervention, proposer_id, created_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			link.ID,
			string(link.RelationshipType),
			link.Rationale.Claim,
			link.Rationale.Scope,
			link.Rationale.Exclusions,
			string(link.State),
			string(link.LeftIntervention),
			string(link.RightIntervention),
			link.ProposerID,
			formatTime(link.CreatedAt),
		)
		if err != nil {
			return fmt.Errorf("insert link: %w", err)
		}

		for _, sectionID := range link.LeftSectionIDs {
			_, err := tx.ExecContext(ctx,
				`INSERT INTO link_sections (link_id, section_id, side) VALUES (?, ?, 'left')`,
				link.ID, sectionID,
			)
			if err != nil {
				return fmt.Errorf("insert left section id: %w", err)
			}
		}
		for _, sectionID := range link.RightSectionIDs {
			_, err := tx.ExecContext(ctx,
				`INSERT INTO link_sections (link_id, section_id, side) VALUES (?, ?, 'right')`,
				link.ID, sectionID,
			)
			if err != nil {
				return fmt.Errorf("insert right section id: %w", err)
			}
		}
		return nil
	})
}

const linkSelectWithSections = `SELECT l.id, l.relationship_type, l.rationale_claim, l.rationale_scope, l.rationale_exclusions, l.state, l.left_intervention, l.right_intervention, l.proposer_id, l.created_at,
	COALESCE(GROUP_CONCAT(CASE WHEN ls.side = 'left' THEN ls.section_id END, ','), '') AS left_ids,
	COALESCE(GROUP_CONCAT(CASE WHEN ls.side = 'right' THEN ls.section_id END, ','), '') AS right_ids
FROM links l
LEFT JOIN link_sections ls ON l.id = ls.link_id`

// FindLinkByID retrieves a link by ID, including its section IDs.
func (r *LinkRepo) FindLinkByID(ctx context.Context, id string) (*core.Link, error) {
	row := r.db.QueryRowContext(ctx, linkSelectWithSections+` WHERE l.id = ? GROUP BY l.id`, id)
	link, err := scanLinkWithSections(row)
	if err == sql.ErrNoRows {
		return nil, core.ErrNotFound{Entity: "link", ID: id}
	}
	if err != nil {
		return nil, fmt.Errorf("find link %s: %w", id, err)
	}
	return link, nil
}

// ListLinks returns all links, optionally filtered by state.
func (r *LinkRepo) ListLinks(ctx context.Context, stateFilter string) ([]*core.Link, error) {
	var query string
	var args []any

	if stateFilter != "" {
		query = linkSelectWithSections + ` WHERE l.state = ? GROUP BY l.id ORDER BY l.created_at`
		args = []any{stateFilter}
	} else {
		query = linkSelectWithSections + ` GROUP BY l.id ORDER BY l.created_at`
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list links: %w", err)
	}
	defer rows.Close()

	var links []*core.Link
	for rows.Next() {
		link, err := scanLinkWithSections(rows)
		if err != nil {
			return nil, err
		}
		links = append(links, link)
	}
	return links, rows.Err()
}

// UpdateLinkState transitions a link's state.
func (r *LinkRepo) UpdateLinkState(ctx context.Context, id string, newState core.LinkState) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE links SET state = ? WHERE id = ?`,
		string(newState), id,
	)
	if err != nil {
		return fmt.Errorf("update link state: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return core.ErrNotFound{Entity: "link", ID: id}
	}
	return nil
}

// AddThreadEntry appends an entry to a link's thread.
func (r *LinkRepo) AddThreadEntry(ctx context.Context, linkID string, entryType core.EntryType, principalID, body string) error {
	id := core.NewID().String()
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO thread_entries (id, link_id, entry_type, principal_id, body) VALUES (?, ?, ?, ?, ?)`,
		id, linkID, string(entryType), principalID, body,
	)
	if err != nil {
		return fmt.Errorf("insert thread entry: %w", err)
	}
	return nil
}

// GetThread returns all thread entries for a link in chronological order.
func (r *LinkRepo) GetThread(ctx context.Context, linkID string) ([]core.ThreadEntry, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, link_id, entry_type, principal_id, body, created_at
		 FROM thread_entries WHERE link_id = ? ORDER BY created_at`,
		linkID,
	)
	if err != nil {
		return nil, fmt.Errorf("query thread: %w", err)
	}
	defer rows.Close()

	var entries []core.ThreadEntry
	for rows.Next() {
		var e core.ThreadEntry
		var entryType, createdAt string
		if err := rows.Scan(&e.ID, &e.LinkID, &entryType, &e.PrincipalID, &e.Body, &createdAt); err != nil {
			return nil, fmt.Errorf("scan thread entry: %w", err)
		}
		e.Type = core.EntryType(entryType)
		e.CreatedAt = parseTime(createdAt)
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// LinksContainingSection returns all links that reference the given section ID on either side.
func (r *LinkRepo) LinksContainingSection(ctx context.Context, sectionID string) ([]*core.Link, error) {
	query := linkSelectWithSections + ` WHERE l.id IN (SELECT link_id FROM link_sections WHERE section_id = ?) GROUP BY l.id ORDER BY l.created_at`
	rows, err := r.db.QueryContext(ctx, query, sectionID)
	if err != nil {
		return nil, fmt.Errorf("links containing section: %w", err)
	}
	defer rows.Close()

	var links []*core.Link
	for rows.Next() {
		link, err := scanLinkWithSections(rows)
		if err != nil {
			return nil, err
		}
		links = append(links, link)
	}
	return links, rows.Err()
}

func scanLinkWithSections(row scannable) (*core.Link, error) {
	var link core.Link
	var relType, state, leftInt, rightInt, createdAt, leftIDsStr, rightIDsStr string
	err := row.Scan(&link.ID, &relType, &link.Rationale.Claim, &link.Rationale.Scope, &link.Rationale.Exclusions,
		&state, &leftInt, &rightInt, &link.ProposerID, &createdAt, &leftIDsStr, &rightIDsStr)
	if err != nil {
		return nil, err
	}
	populateLink(&link, relType, state, leftInt, rightInt, createdAt, leftIDsStr, rightIDsStr)
	return &link, nil
}

// LinksContainingSections returns a map from section ID to the links containing that section.
func (r *LinkRepo) LinksContainingSections(ctx context.Context, sectionIDs []string) (map[string][]*core.LinkInfo, error) {
	if len(sectionIDs) == 0 {
		return nil, nil
	}
	placeholders := make([]string, len(sectionIDs))
	args := make([]any, len(sectionIDs))
	for i, id := range sectionIDs {
		placeholders[i] = "?"
		args[i] = id
	}
	query := `SELECT DISTINCT l.id, l.state, l.relationship_type,
		COALESCE(GROUP_CONCAT(CASE WHEN ls.side = 'left' THEN ls.section_id END, ','), ''),
		COALESCE(GROUP_CONCAT(CASE WHEN ls.side = 'right' THEN ls.section_id END, ','), '')
		FROM links l
		JOIN link_sections ls ON l.id = ls.link_id
		WHERE l.id IN (SELECT DISTINCT link_id FROM link_sections WHERE section_id IN (` + strings.Join(placeholders, ",") + `))
		GROUP BY l.id`
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("links containing sections batch: %w", err)
	}
	defer rows.Close()

	// Collect all links first.
	var links []*core.LinkInfo
	for rows.Next() {
		var li core.LinkInfo
		var leftIDsStr, rightIDsStr string
		if err := rows.Scan(&li.ID, &li.State, &li.RelationshipType, &leftIDsStr, &rightIDsStr); err != nil {
			return nil, fmt.Errorf("scan link info: %w", err)
		}
		if leftIDsStr != "" {
			li.LeftSectionIDs = strings.Split(leftIDsStr, ",")
		}
		if rightIDsStr != "" {
			li.RightSectionIDs = strings.Split(rightIDsStr, ",")
		}
		links = append(links, &li)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Build map: section ID -> links containing that section.
	requested := make(map[string]bool, len(sectionIDs))
	for _, id := range sectionIDs {
		requested[id] = true
	}
	result := make(map[string][]*core.LinkInfo)
	for _, li := range links {
		for _, sid := range li.LeftSectionIDs {
			if requested[sid] {
				result[sid] = append(result[sid], li)
			}
		}
		for _, sid := range li.RightSectionIDs {
			if requested[sid] {
				result[sid] = append(result[sid], li)
			}
		}
	}
	return result, nil
}

func populateLink(link *core.Link, relType, state, leftInt, rightInt, createdAt, leftIDsStr, rightIDsStr string) {
	link.RelationshipType = core.RelationshipType(relType)
	link.State = core.LinkState(state)
	link.LeftIntervention = core.InterventionLevel(leftInt)
	link.RightIntervention = core.InterventionLevel(rightInt)
	link.CreatedAt = parseTime(createdAt)

	if leftIDsStr != "" {
		link.LeftSectionIDs = strings.Split(leftIDsStr, ",")
	}
	if rightIDsStr != "" {
		link.RightSectionIDs = strings.Split(rightIDsStr, ",")
	}
}
