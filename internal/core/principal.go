package core

// PrincipalType distinguishes human users from automated services.
type PrincipalType string

const (
	PrincipalHuman   PrincipalType = "human"
	PrincipalService PrincipalType = "service"
)

// Principal represents the actor performing an action.
type Principal struct {
	ID   string
	Type PrincipalType
	Name string
}

// RequireHuman returns ErrUnauthorized if the principal is not human.
func (p Principal) RequireHuman(action string) error {
	if p.Type != PrincipalHuman {
		return ErrUnauthorized{Action: action, PrincipalType: string(p.Type)}
	}
	return nil
}
