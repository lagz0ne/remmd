package playbook

import "testing"

func TestPrincipal_HasTeam(t *testing.T) {
	t.Parallel()
	p := Principal{ID: "agent-1", Type: "service", Teams: []string{"backend", "platform"}, Roles: []string{"engineer"}}
	if !p.HasTeam("backend") {
		t.Fatal("expected HasTeam(backend) = true")
	}
	if p.HasTeam("frontend") {
		t.Fatal("expected HasTeam(frontend) = false")
	}
}

func TestPrincipal_HasRole(t *testing.T) {
	t.Parallel()
	p := Principal{ID: "user-1", Type: "human", Teams: []string{"platform"}, Roles: []string{"admin", "reviewer"}}
	if !p.HasRole("admin") {
		t.Fatal("expected HasRole(admin) = true")
	}
	if p.HasRole("engineer") {
		t.Fatal("expected HasRole(engineer) = false")
	}
}

func TestPrincipal_ToMap(t *testing.T) {
	t.Parallel()
	p := Principal{ID: "agent-1", Type: "service", Teams: []string{"backend"}, Roles: []string{"engineer"}}
	m := p.ToMap()
	if m["id"] != "agent-1" || m["type"] != "service" {
		t.Fatalf("unexpected map: %v", m)
	}
	teams, ok := m["teams"].([]any)
	if !ok || len(teams) != 1 || teams[0] != "backend" {
		t.Fatalf("teams: %v", m["teams"])
	}
}
