package main

import (
	"testing"
	"time"
)

func TestAuthz_Evaluate_AllowRole(t *testing.T) {
	eng := NewAuthzEngine()
	claims := Claims{
		Iss:   "https://idp.example.invalid",
		Sub:   "user-1",
		Aud:   "eng-h-004",
		Exp:   time.Now().UTC().Add(time.Hour).Unix(),
		Roles: []string{"developer", "viewer"},
	}
	decision := eng.Evaluate(claims, "pipelines", "read")
	if !decision.Allow {
		t.Fatalf("expected allow: %+v", decision)
	}
	if decision.Reason != "rbac_allow" {
		t.Fatalf("reason=%q want rbac_allow", decision.Reason)
	}
}

func TestAuthz_Evaluate_DenyMissingRole(t *testing.T) {
	eng := NewAuthzEngine()
	claims := Claims{
		Iss:   "https://idp.example.invalid",
		Sub:   "user-2",
		Aud:   "eng-h-004",
		Exp:   time.Now().UTC().Add(time.Hour).Unix(),
		Roles: []string{"viewer"},
	}
	decision := eng.Evaluate(claims, "pipelines", "write")
	if decision.Allow {
		t.Fatalf("expected deny: %+v", decision)
	}
	if decision.Reason != "rbac_deny" {
		t.Fatalf("reason=%q want rbac_deny", decision.Reason)
	}
}

func TestAuthz_Claims_ValidateExp(t *testing.T) {
	eng := NewAuthzEngine()
	claims := Claims{
		Iss:   "https://idp.example.invalid",
		Sub:   "user-3",
		Aud:   "eng-h-004",
		Exp:   time.Now().UTC().Add(-time.Minute).Unix(),
		Roles: []string{"admin"},
	}
	decision := eng.Evaluate(claims, "pipelines", "write")
	if decision.Allow {
		t.Fatalf("expired claims must deny: %+v", decision)
	}
}

func TestAuthz_Info_OIDCInspired(t *testing.T) {
	eng := NewAuthzEngine()
	info := eng.Info()
	if info["oidc_inspired"] != true {
		t.Fatalf("oidc_inspired want true: %+v", info)
	}
	if info["simulator"] != true {
		t.Fatalf("simulator want true: %+v", info)
	}
	if info["external_idp"] == true {
		t.Fatal("must not claim external IdP")
	}
}
