package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// Org mirrors one entry in `shopify organization list -j`.
type Org struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Store mirrors one entry in `shopify store list -j`.
type Store struct {
	Name   string `json:"name"`  // friendly name, e.g. "doe-test-martin"
	Domain string `json:"store"` // full domain, e.g. doe-test-martin.myshopify.com
	Type   string `json:"type"`  // dev | production | client-transfer | collaborator
}

// These mirror the *wrapper objects* the JSON actually ships in.
type orgList struct {
	Organizations []Org `json:"organizations"`
}

type storeList struct {
	Stores []Store `json:"stores"`
}

// cliError is the envelope `shopify` prints on stdout when a command fails.
// The CLI also exits non-zero, so without this the only clue is "exit status 2".
type cliError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func (e *cliError) Error() string {
	if e.Type == "" {
		return e.Message
	}
	return e.Type + ": " + e.Message
}

// runShopifyJSON runs `shopify <args>` and returns its stdout.
func runShopifyJSON(args ...string) ([]byte, error) {
	var stdout bytes.Buffer
	cmd := exec.Command("shopify", args...)
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		// The reason for the failure lives in stdout, not stderr.
		var env struct {
			Error *cliError `json:"error"`
		}
		if json.Unmarshal(stdout.Bytes(), &env) == nil && env.Error != nil {
			return nil, fmt.Errorf("shopify %s: %w", strings.Join(args, " "), env.Error)
		}
		return nil, fmt.Errorf("shopify %s: %w", strings.Join(args, " "), err)
	}
	return stdout.Bytes(), nil
}

// snippet trims a payload to something safe to quote inside an error.
func snippet(raw []byte) string {
	s := strings.TrimSpace(string(raw))
	if len(s) > 200 {
		return s[:200] + "…"
	}
	return s
}

// fetchCLI runs `shopify <args> -j` and decodes stdout into v. Decode failures
// quote the payload: the CLI occasionally emits a notice *after* the JSON,
// which otherwise surfaces as a bare "invalid character 'c' after top-level
// value" with no hint as to which command or what text.
func fetchCLI(args []string, v any) error {
	raw, err := runShopifyJSON(args...)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(raw, v); err != nil {
		return fmt.Errorf("shopify %s: %w (payload: %s)", strings.Join(args, " "), err, snippet(raw))
	}
	return nil
}

func fetchOrgs() ([]Org, error) {
	var res orgList
	if err := fetchCLI([]string{"organization", "list", "-j"}, &res); err != nil {
		return nil, err
	}
	return res.Organizations, nil
}

func fetchStores(orgID string) ([]Store, error) {
	var res storeList
	if err := fetchCLI([]string{"store", "list", "-j", "--organization-id", orgID}, &res); err != nil {
		return nil, err
	}
	return res.Stores, nil
}
