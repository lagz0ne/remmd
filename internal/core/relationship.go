package core

// RelationshipType classifies the semantic relationship between linked section groups.
type RelationshipType string

const (
	RelAgreesWith  RelationshipType = "agrees_with"
	RelImplements  RelationshipType = "implements"
	RelTests       RelationshipType = "tests"
	RelEvidences   RelationshipType = "evidences"
)

// ValidRelationshipType returns true if s is a recognized relationship type.
func ValidRelationshipType(s string) bool {
	switch RelationshipType(s) {
	case RelAgreesWith, RelImplements, RelTests, RelEvidences:
		return true
	}
	return false
}
