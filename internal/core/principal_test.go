package core_test

import (
	"errors"
	"testing"

	"github.com/lagz0ne/remmd/internal/core"
)

func TestPrincipal_HumanRequireHuman_ReturnsNil(t *testing.T) {
	t.Parallel()
	p := core.Principal{ID: "u1", Type: core.PrincipalHuman, Name: "Alice"}
	if err := p.RequireHuman("approve"); err != nil {
		t.Errorf("RequireHuman() = %v, want nil", err)
	}
}

func TestPrincipal_ServiceRequireHuman_ReturnsErrUnauthorized(t *testing.T) {
	t.Parallel()
	p := core.Principal{ID: "svc1", Type: core.PrincipalService, Name: "bot"}
	err := p.RequireHuman("approve")
	if err == nil {
		t.Fatal("RequireHuman() = nil, want ErrUnauthorized")
	}
	var unauth core.ErrUnauthorized
	if !errors.As(err, &unauth) {
		t.Fatalf("error type = %T, want core.ErrUnauthorized", err)
	}
}

func TestPrincipal_ErrUnauthorized_HasCorrectFields(t *testing.T) {
	t.Parallel()
	p := core.Principal{ID: "svc1", Type: core.PrincipalService, Name: "bot"}
	err := p.RequireHuman("withdraw")
	var unauth core.ErrUnauthorized
	if !errors.As(err, &unauth) {
		t.Fatalf("error type = %T, want core.ErrUnauthorized", err)
	}
	if unauth.Action != "withdraw" {
		t.Errorf("Action = %q, want %q", unauth.Action, "withdraw")
	}
	if unauth.PrincipalType != "service" {
		t.Errorf("PrincipalType = %q, want %q", unauth.PrincipalType, "service")
	}
}
