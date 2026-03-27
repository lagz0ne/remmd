package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lagz0ne/remmd/internal/core"
	"github.com/lagz0ne/remmd/internal/playbook"
)

type PlaybookStore struct {
	db *sql.DB
}

func NewPlaybookStore(db *sql.DB) *PlaybookStore {
	return &PlaybookStore{db: db}
}

func (s *PlaybookStore) Import(ctx context.Context, name string, yamlData []byte) (version int, isNew bool, err error) {
	pb, err := playbook.Parse(yamlData)
	if err != nil {
		return 0, false, fmt.Errorf("parse playbook: %w", err)
	}

	hash := sha256hex(yamlData)

	var latestVer int
	var latestHash string
	row := s.db.QueryRowContext(ctx,
		`SELECT version, hash FROM playbooks WHERE name = ? ORDER BY version DESC LIMIT 1`, name)
	if err := row.Scan(&latestVer, &latestHash); err != nil && err != sql.ErrNoRows {
		return 0, false, fmt.Errorf("query latest version: %w", err)
	}

	if latestHash == hash {
		return latestVer, false, nil
	}

	newVer := latestVer + 1
	pbID := core.NewID().String()

	err = WithTx(ctx, s.db, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO playbooks (id, name, version, hash) VALUES (?, ?, ?, ?)`,
			pbID, name, newVer, hash); err != nil {
			return fmt.Errorf("insert playbook: %w", err)
		}

		for typeName, td := range pb.Types {
			typeID, err := insertType(ctx, tx, pbID, typeName, td)
			if err != nil {
				return err
			}

			for fieldName, fd := range td.Fields {
				if _, err := tx.ExecContext(ctx,
					`INSERT INTO pb_fields (type_id, name, notation) VALUES (?, ?, ?)`,
					typeID, fieldName, fieldNotation(fd)); err != nil {
					return fmt.Errorf("insert field %q.%q: %w", typeName, fieldName, err)
				}
			}

			for i, sd := range td.Sections {
				if _, err := tx.ExecContext(ctx,
					`INSERT INTO pb_sections (type_id, name, required, "order") VALUES (?, ?, ?, ?)`,
					typeID, sd.Name, sd.Required, i); err != nil {
					return fmt.Errorf("insert section %q.%q: %w", typeName, sd.Name, err)
				}
			}

			for ruleName, rd := range td.Rules {
				if err := insertRule(ctx, tx, pbID, typeName, ruleName, rd); err != nil {
					return err
				}
			}
		}

		for edgeName, ed := range pb.Edges {
			if _, err := tx.ExecContext(ctx,
				`INSERT INTO pb_edges (playbook_id, name, notation) VALUES (?, ?, ?)`,
				pbID, edgeName, edgeNotation(ed)); err != nil {
				return fmt.Errorf("insert edge %q: %w", edgeName, err)
			}
		}

		for ruleName, rd := range pb.Rules {
			if err := insertRule(ctx, tx, pbID, "", ruleName, rd); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return 0, false, err
	}

	return newVer, true, nil
}

func (s *PlaybookStore) Latest(ctx context.Context, name string) (*playbook.Playbook, int, error) {
	var pbID string
	var version int
	row := s.db.QueryRowContext(ctx,
		`SELECT id, version FROM playbooks WHERE name = ? ORDER BY version DESC LIMIT 1`, name)
	if err := row.Scan(&pbID, &version); err != nil {
		if err == sql.ErrNoRows {
			return nil, 0, nil
		}
		return nil, 0, fmt.Errorf("query latest playbook: %w", err)
	}

	pb := &playbook.Playbook{
		Types: make(map[string]*playbook.TypeDef),
		Edges: make(map[string]*playbook.EdgeDef),
		Rules: make(map[string]*playbook.RuleDef),
	}

	typeRows, err := s.db.QueryContext(ctx,
		`SELECT id, name, description FROM pb_types WHERE playbook_id = ?`, pbID)
	if err != nil {
		return nil, 0, fmt.Errorf("query types: %w", err)
	}
	defer typeRows.Close()

	type typeRow struct {
		id   int64
		name string
		desc string
	}
	var types []typeRow
	for typeRows.Next() {
		var tr typeRow
		if err := typeRows.Scan(&tr.id, &tr.name, &tr.desc); err != nil {
			return nil, 0, fmt.Errorf("scan type: %w", err)
		}
		types = append(types, tr)
	}
	if err := typeRows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate types: %w", err)
	}

	for _, tr := range types {
		td := &playbook.TypeDef{
			Name:        tr.name,
			Description: tr.desc,
			Fields:      make(map[string]playbook.FieldDef),
			Rules:       make(map[string]*playbook.RuleDef),
		}

		fieldRows, err := s.db.QueryContext(ctx,
			`SELECT name, notation FROM pb_fields WHERE type_id = ?`, tr.id)
		if err != nil {
			return nil, 0, fmt.Errorf("query fields for %q: %w", tr.name, err)
		}
		for fieldRows.Next() {
			var fname, notation string
			if err := fieldRows.Scan(&fname, &notation); err != nil {
				fieldRows.Close()
				return nil, 0, fmt.Errorf("scan field: %w", err)
			}
			fd, err := playbook.ParseField(notation)
			if err != nil {
				fieldRows.Close()
				return nil, 0, fmt.Errorf("parse field notation %q: %w", notation, err)
			}
			td.Fields[fname] = fd
		}
		fieldRows.Close()

		secRows, err := s.db.QueryContext(ctx,
			`SELECT name, required FROM pb_sections WHERE type_id = ? ORDER BY "order"`, tr.id)
		if err != nil {
			return nil, 0, fmt.Errorf("query sections for %q: %w", tr.name, err)
		}
		for secRows.Next() {
			var sd playbook.SectionDef
			if err := secRows.Scan(&sd.Name, &sd.Required); err != nil {
				secRows.Close()
				return nil, 0, fmt.Errorf("scan section: %w", err)
			}
			td.Sections = append(td.Sections, sd)
		}
		secRows.Close()

		pb.Types[tr.name] = td
	}

	ruleRows, err := s.db.QueryContext(ctx,
		`SELECT id, scope_type, name, description, severity, expr FROM pb_rules WHERE playbook_id = ?`, pbID)
	if err != nil {
		return nil, 0, fmt.Errorf("query rules: %w", err)
	}
	defer ruleRows.Close()

	type ruleRow struct {
		id        int64
		scopeType string
		name      string
		desc      string
		severity  string
		expr      string
	}
	var rules []ruleRow
	for ruleRows.Next() {
		var rr ruleRow
		if err := ruleRows.Scan(&rr.id, &rr.scopeType, &rr.name, &rr.desc, &rr.severity, &rr.expr); err != nil {
			return nil, 0, fmt.Errorf("scan rule: %w", err)
		}
		rules = append(rules, rr)
	}
	if err := ruleRows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate rules: %w", err)
	}

	for _, rr := range rules {
		rd := &playbook.RuleDef{
			Name:        rr.name,
			Description: rr.desc,
			Severity:    rr.severity,
			Expr:        rr.expr,
		}

		exRows, err := s.db.QueryContext(ctx,
			`SELECT pass, data_json FROM pb_examples WHERE rule_id = ? ORDER BY "order"`, rr.id)
		if err != nil {
			return nil, 0, fmt.Errorf("query examples for rule %q: %w", rr.name, err)
		}
		for exRows.Next() {
			var ex playbook.Example
			var dataJSON string
			if err := exRows.Scan(&ex.Pass, &dataJSON); err != nil {
				exRows.Close()
				return nil, 0, fmt.Errorf("scan example: %w", err)
			}
			if err := json.Unmarshal([]byte(dataJSON), &ex.Data); err != nil {
				exRows.Close()
				return nil, 0, fmt.Errorf("unmarshal example data: %w", err)
			}
			rd.Examples = append(rd.Examples, ex)
		}
		exRows.Close()

		if rr.scopeType == "" {
			pb.Rules[rr.name] = rd
		} else {
			if td, ok := pb.Types[rr.scopeType]; ok {
				td.Rules[rr.name] = rd
			}
		}
	}

	edgeRows, err := s.db.QueryContext(ctx,
		`SELECT name, notation FROM pb_edges WHERE playbook_id = ?`, pbID)
	if err != nil {
		return nil, 0, fmt.Errorf("query edges: %w", err)
	}
	defer edgeRows.Close()

	for edgeRows.Next() {
		var ename, notation string
		if err := edgeRows.Scan(&ename, &notation); err != nil {
			return nil, 0, fmt.Errorf("scan edge: %w", err)
		}
		ed, err := playbook.ParseEdge(notation)
		if err != nil {
			return nil, 0, fmt.Errorf("parse edge notation %q: %w", notation, err)
		}
		ed.Name = ename
		pb.Edges[ename] = &ed
	}

	return pb, version, edgeRows.Err()
}

func (s *PlaybookStore) LatestVersion(ctx context.Context, name string) (version int, hash string, err error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT version, hash FROM playbooks WHERE name = ? ORDER BY version DESC LIMIT 1`, name)
	if err := row.Scan(&version, &hash); err != nil {
		if err == sql.ErrNoRows {
			return 0, "", nil
		}
		return 0, "", fmt.Errorf("query latest version: %w", err)
	}
	return version, hash, nil
}

func insertType(ctx context.Context, tx *sql.Tx, pbID, name string, td *playbook.TypeDef) (int64, error) {
	res, err := tx.ExecContext(ctx,
		`INSERT INTO pb_types (playbook_id, name, description) VALUES (?, ?, ?)`,
		pbID, name, td.Description)
	if err != nil {
		return 0, fmt.Errorf("insert type %q: %w", name, err)
	}
	return res.LastInsertId()
}

func insertRule(ctx context.Context, tx *sql.Tx, pbID, scopeType, name string, rd *playbook.RuleDef) error {
	res, err := tx.ExecContext(ctx,
		`INSERT INTO pb_rules (playbook_id, scope_type, name, description, severity, expr) VALUES (?, ?, ?, ?, ?, ?)`,
		pbID, scopeType, name, rd.Description, rd.Severity, rd.Expr)
	if err != nil {
		return fmt.Errorf("insert rule %q: %w", name, err)
	}

	if len(rd.Examples) > 0 {
		ruleID, err := res.LastInsertId()
		if err != nil {
			return fmt.Errorf("last insert id for rule %q: %w", name, err)
		}
		for i, ex := range rd.Examples {
			dataJSON, err := json.Marshal(ex.Data)
			if err != nil {
				return fmt.Errorf("marshal example data: %w", err)
			}
			if _, err := tx.ExecContext(ctx,
				`INSERT INTO pb_examples (rule_id, pass, data_json, "order") VALUES (?, ?, ?, ?)`,
				ruleID, ex.Pass, string(dataJSON), i); err != nil {
				return fmt.Errorf("insert example for rule %q: %w", name, err)
			}
		}
	}

	return nil
}

func sha256hex(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func fieldNotation(fd playbook.FieldDef) string {
	var b strings.Builder
	b.WriteString(fd.Type)
	switch fd.Type {
	case "enum":
		b.WriteByte('(')
		b.WriteString(strings.Join(fd.Values, ", "))
		b.WriteByte(')')
	case "ref":
		b.WriteByte('(')
		if len(fd.Targets) > 0 {
			b.WriteString(strings.Join(fd.Targets, " | "))
		} else if fd.Target != "" {
			b.WriteString(fd.Target)
		}
		b.WriteByte(')')
	case "list":
		if fd.Target != "" {
			b.WriteByte('(')
			b.WriteString(fd.Target)
			b.WriteByte(')')
		}
	}
	if fd.Required {
		b.WriteByte('!')
	}
	return b.String()
}

func edgeNotation(ed *playbook.EdgeDef) string {
	max := "*"
	if ed.MaxCard >= 0 {
		max = fmt.Sprintf("%d", ed.MaxCard)
	}
	return fmt.Sprintf("%s -> %s [%d..%s]", strings.Join(ed.From, " | "), ed.To, ed.MinCard, max)
}
