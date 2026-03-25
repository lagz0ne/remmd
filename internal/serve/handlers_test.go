package serve

import (
	"testing"

	"github.com/lagz0ne/remmd/internal/core"
)

func TestSubjectPart(t *testing.T) {
	t.Parallel()
	tests := []struct {
		subject string
		index   int
		want    string
		wantErr bool
	}{
		{"remmd.q.documents.d1.sections", 3, "d1", false},
		{"remmd.q.section.a1", 3, "a1", false},
		{"remmd.q.documents", 0, "remmd", false},
		{"remmd.q", 3, "", true},
		{"a", 1, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.subject, func(t *testing.T) {
			got, err := subjectPart(tt.subject, tt.index)
			if (err != nil) != tt.wantErr {
				t.Fatalf("subjectPart(%q, %d) error = %v, wantErr %v", tt.subject, tt.index, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("subjectPart(%q, %d) = %q, want %q", tt.subject, tt.index, got, tt.want)
			}
		})
	}
}

func TestDeriveLinkInfoState_Priority(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		states []string
		want   string
	}{
		{"no links", nil, ""},
		{"single aligned", []string{"aligned"}, "aligned"},
		{"stale + broken = broken", []string{"stale", "broken"}, "broken"},
		{"archived only = empty", []string{"archived"}, ""},
		{"archived + pending = pending", []string{"archived", "pending"}, "pending"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var links []*core.LinkInfo
			for _, s := range tt.states {
				links = append(links, &core.LinkInfo{State: s})
			}
			got := deriveLinkInfoState(links)
			if got != tt.want {
				t.Errorf("deriveLinkInfoState(%v) = %q, want %q", tt.states, got, tt.want)
			}
		})
	}
}

func TestDeriveLinkState_Priority(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		states []core.LinkState
		want   string
	}{
		{"no links", nil, ""},
		{"single aligned", []core.LinkState{core.LinkAligned}, "aligned"},
		{"single pending", []core.LinkState{core.LinkPending}, "pending"},
		{"single stale", []core.LinkState{core.LinkStale}, "stale"},
		{"single broken", []core.LinkState{core.LinkBroken}, "broken"},
		{"aligned + pending = pending", []core.LinkState{core.LinkAligned, core.LinkPending}, "pending"},
		{"aligned + stale = stale", []core.LinkState{core.LinkAligned, core.LinkStale}, "stale"},
		{"pending + stale = stale", []core.LinkState{core.LinkPending, core.LinkStale}, "stale"},
		{"stale + broken = broken", []core.LinkState{core.LinkStale, core.LinkBroken}, "broken"},
		{"all states = broken", []core.LinkState{core.LinkAligned, core.LinkPending, core.LinkStale, core.LinkBroken}, "broken"},
		{"archived only = empty", []core.LinkState{core.LinkArchived}, ""},
		{"archived + aligned = aligned", []core.LinkState{core.LinkArchived, core.LinkAligned}, "aligned"},
		{"archived + broken = broken", []core.LinkState{core.LinkArchived, core.LinkBroken}, "broken"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var links []*core.Link
			for _, s := range tt.states {
				links = append(links, &core.Link{State: s})
			}
			got := deriveLinkState(links)
			if got != tt.want {
				t.Errorf("deriveLinkState(%v) = %q, want %q", tt.states, got, tt.want)
			}
		})
	}
}
