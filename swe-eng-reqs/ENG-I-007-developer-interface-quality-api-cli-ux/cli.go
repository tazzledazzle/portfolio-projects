package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// StatusJSON is the stable CLI status payload for --json output.
type StatusJSON struct {
	OK            bool   `json:"ok"`
	RequirementID string `json:"requirement_id"`
	Service       string `json:"service"`
	Status        string `json:"status"`
}

// RunCLI executes eng-i-007 CLI subcommands. Returns stdout and a clear error
// without panicking or leaking secrets (T-5-04).
func RunCLI(args []string) (string, error) {
	if len(args) == 0 {
		return "", errors.New("usage: eng-i-007 <command> [flags]\ncommands: status")
	}
	cmd := args[0]
	rest := args[1:]
	switch cmd {
	case "status":
		return runStatus(rest)
	default:
		return "", fmt.Errorf("unknown command %q: expected status (see --help)", cmd)
	}
}

func runStatus(args []string) (string, error) {
	jsonOut := false
	for _, a := range args {
		switch a {
		case "--json":
			jsonOut = true
		case "--help", "-h":
			return "", errors.New("usage: eng-i-007 status --json")
		default:
			return "", fmt.Errorf("invalid flag %q: usage: eng-i-007 status --json", a)
		}
	}
	if !jsonOut {
		return "", errors.New("missing required flag --json: usage: eng-i-007 status --json")
	}
	payload := StatusJSON{
		OK:            true,
		RequirementID: "ENG-I-007",
		Service:       "eng-i-007",
		Status:        "ready",
	}
	b, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return "", errors.New("failed to encode status json")
	}
	out := string(b) + "\n"
	// Ensure no accidental secret-like tokens in stdout.
	if strings.Contains(strings.ToLower(out), "password") || strings.Contains(out, "SECRET") {
		return "", errors.New("refusing to emit status that looks secret-bearing")
	}
	return out, nil
}

func normalizeJSONWhitespace(s string) string {
	var v any
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return strings.TrimSpace(s)
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return strings.TrimSpace(s)
	}
	return string(b) + "\n"
}
