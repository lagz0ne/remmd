package core_test

import (
	"testing"

	"github.com/lagz0ne/remmd/internal/core"
)

func TestAgreementSnapshot_SameHashes_SameResult(t *testing.T) {
	t.Parallel()
	s1 := core.AgreementSnapshot{
		LinkID:             "l1",
		LeftContentHashes:  []string{"aaa", "bbb"},
		RightContentHashes: []string{"ccc"},
	}
	s2 := core.AgreementSnapshot{
		LinkID:             "l1",
		LeftContentHashes:  []string{"aaa", "bbb"},
		RightContentHashes: []string{"ccc"},
	}
	if s1.Hash() != s2.Hash() {
		t.Errorf("same inputs should produce same hash: %q != %q", s1.Hash(), s2.Hash())
	}
}

func TestAgreementSnapshot_DifferentHashes_DifferentResult(t *testing.T) {
	t.Parallel()
	s1 := core.AgreementSnapshot{
		LinkID:             "l1",
		LeftContentHashes:  []string{"aaa"},
		RightContentHashes: []string{"ccc"},
	}
	s2 := core.AgreementSnapshot{
		LinkID:             "l1",
		LeftContentHashes:  []string{"bbb"},
		RightContentHashes: []string{"ccc"},
	}
	if s1.Hash() == s2.Hash() {
		t.Error("different inputs should produce different hashes")
	}
}

func TestAgreementSnapshot_HashDeterministic(t *testing.T) {
	t.Parallel()
	s := core.AgreementSnapshot{
		LinkID:             "l1",
		LeftContentHashes:  []string{"aaa", "bbb"},
		RightContentHashes: []string{"ccc"},
	}
	h1 := s.Hash()
	h2 := s.Hash()
	if h1 != h2 {
		t.Errorf("Hash() not deterministic: %q != %q", h1, h2)
	}
}

func TestAgreementSnapshot_Empty_HasDefinedHash(t *testing.T) {
	t.Parallel()
	s := core.AgreementSnapshot{}
	h := s.Hash()
	if h == "" {
		t.Error("empty snapshot should have a defined (non-empty) hash")
	}
}
