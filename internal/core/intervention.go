package core

// InterventionLevel controls how aggressively a side of a link alerts on state changes.
type InterventionLevel string

const (
	InterventionWatch    InterventionLevel = "watch"
	InterventionNotify   InterventionLevel = "notify"
	InterventionUrgent   InterventionLevel = "urgent"
	InterventionBlocking InterventionLevel = "blocking"
)

// ValidInterventionLevel returns true if s is a recognized intervention level.
func ValidInterventionLevel(s string) bool {
	switch InterventionLevel(s) {
	case InterventionWatch, InterventionNotify, InterventionUrgent, InterventionBlocking:
		return true
	}
	return false
}
