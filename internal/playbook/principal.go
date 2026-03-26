package playbook

type Principal struct {
	ID    string
	Type  string   // "human" or "service"
	Teams []string
	Roles []string
}

func (p Principal) HasTeam(team string) bool {
	for _, t := range p.Teams {
		if t == team {
			return true
		}
	}
	return false
}

func (p Principal) HasRole(role string) bool {
	for _, r := range p.Roles {
		if r == role {
			return true
		}
	}
	return false
}

func (p Principal) ToMap() map[string]any {
	teams := make([]any, len(p.Teams))
	for i, t := range p.Teams {
		teams[i] = t
	}
	roles := make([]any, len(p.Roles))
	for i, r := range p.Roles {
		roles[i] = r
	}
	return map[string]any{
		"id":    p.ID,
		"type":  p.Type,
		"teams": teams,
		"roles": roles,
	}
}
