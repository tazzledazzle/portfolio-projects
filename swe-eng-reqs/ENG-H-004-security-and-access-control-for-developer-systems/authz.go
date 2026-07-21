package main

import (
	"strings"
	"sync"
	"time"
)

// Claims is an OIDC-inspired claim set (simulator — no external IdP).
type Claims struct {
	Iss   string   `json:"iss"`
	Sub   string   `json:"sub"`
	Aud   string   `json:"aud"`
	Exp   int64    `json:"exp"`
	Roles []string `json:"roles"`
}

// Decision is the RBAC evaluation result.
type Decision struct {
	Allow  bool   `json:"allow"`
	Reason string `json:"reason"`
	Action string `json:"action"`
	Resource string `json:"resource"`
}

// policy maps resource/action → required role (default deny).
var defaultPolicies = map[string]string{
	"pipelines:read":  "viewer",
	"pipelines:write": "developer",
	"admin:write":     "admin",
}

// AuthzEngine evaluates OIDC-inspired claims against RBAC policies.
type AuthzEngine struct {
	mu       sync.Mutex
	policies map[string]string
	nowFn    func() time.Time
}

// NewAuthzEngine returns a default-deny RBAC engine with OIDC-inspired claim checks.
func NewAuthzEngine() *AuthzEngine {
	policies := make(map[string]string, len(defaultPolicies))
	for k, v := range defaultPolicies {
		policies[k] = v
	}
	return &AuthzEngine{
		policies: policies,
		nowFn:    func() time.Time { return time.Now().UTC() },
	}
}

// Evaluate validates exp then applies RBAC (T-5-19).
func (e *AuthzEngine) Evaluate(claims Claims, resource, action string) Decision {
	resource = strings.TrimSpace(resource)
	action = strings.TrimSpace(action)
	if claims.Sub == "" || claims.Iss == "" || claims.Aud == "" {
		return Decision{Allow: false, Reason: "rbac_deny", Resource: resource, Action: action}
	}
	now := e.nowFn().Unix()
	if claims.Exp <= now {
		return Decision{Allow: false, Reason: "rbac_deny", Resource: resource, Action: action}
	}

	e.mu.Lock()
	required, ok := e.policies[resource+":"+action]
	e.mu.Unlock()
	if !ok {
		return Decision{Allow: false, Reason: "rbac_deny", Resource: resource, Action: action}
	}
	if hasRole(claims.Roles, required) || hasRole(claims.Roles, "admin") {
		return Decision{Allow: true, Reason: "rbac_allow", Resource: resource, Action: action}
	}
	return Decision{Allow: false, Reason: "rbac_deny", Resource: resource, Action: action}
}

// Info returns honesty labels — oidc_inspired simulator, no external IdP.
func (e *AuthzEngine) Info() map[string]any {
	return map[string]any{
		"requirement_id": "ENG-H-004",
		"service":        "eng-h-004",
		"oidc_inspired":  true,
		"simulator":      true,
		"external_idp":   false,
		"note":           "OIDC-inspired claims (iss/sub/aud/exp/roles) + RBAC; no Keycloak/Dex/OIDC client",
	}
}

func hasRole(roles []string, want string) bool {
	for _, role := range roles {
		if role == want {
			return true
		}
	}
	return false
}
