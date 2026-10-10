package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestFetchCLIDecodesOrgList(t *testing.T) {
	raw := []byte(`{"organizations":[{"id":"188480102","name":"Martin Jose","shopCount":1}]}`)
	var got orgList
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(got.Organizations) != 1 || got.Organizations[0].ID != "188480102" {
		t.Fatalf("got %+v", got.Organizations)
	}
}

func TestFetchCLIDecodesStoreList(t *testing.T) {
	raw := []byte(`{"stores":[{"store":"a.myshopify.com","name":"a","type":"dev"}]}`)
	var got storeList
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(got.Stores) != 1 || got.Stores[0].Domain != "a.myshopify.com" {
		t.Fatalf("got %+v", got.Stores)
	}
}

// The CLI has emitted a notice *after* valid JSON, which used to surface as a
// bare "invalid character 'c' after top-level value". The error must name the
// command and quote the payload.
func TestTrailingGarbageIsDiagnosable(t *testing.T) {
	payload := `{"organizations":[]}` + "\ncommand not found\n"
	err := json.Unmarshal([]byte(payload), &orgList{})
	if err == nil {
		t.Fatal("expected a decode error")
	}
	if !strings.Contains(err.Error(), "after top-level value") {
		t.Errorf("error lost the decode detail: %v", err)
	}

	if got := snippet([]byte(payload)); !strings.Contains(got, "command not found") {
		t.Errorf("snippet did not preserve the offending text: %q", got)
	}
}

func TestSnippetTruncates(t *testing.T) {
	long := strings.Repeat("x", 500)
	got := snippet([]byte(long))
	if len([]rune(got)) != 201 {
		t.Errorf("want 200 runes + ellipsis, got %d", len([]rune(got)))
	}
	if !strings.HasSuffix(got, "…") {
		t.Errorf("want ellipsis suffix, got %q", got[len(got)-4:])
	}
}

func TestSnippetTrimsWhitespace(t *testing.T) {
	if got := snippet([]byte("  {\"a\":1}\n")); got != `{"a":1}` {
		t.Errorf("got %q", got)
	}
}

// A failing CLI command reports its reason as JSON on stdout and exits
// non-zero; that reason must beat "exit status 2".
func TestCLIErrorMessage(t *testing.T) {
	tests := []struct {
		name string
		env  cliError
		want string
	}{
		{"typed", cliError{Type: "abort", Message: "An organization ID is required"}, "abort: An organization ID is required"},
		{"untyped", cliError{Message: "boom"}, "boom"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.env.Error(); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCLIErrorEnvelopeParses(t *testing.T) {
	raw := []byte(`{"error":{"type":"abort","message":"Nonexistent flag: --nope"}}`)
	var env struct {
		Error *cliError `json:"error"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if env.Error == nil {
		t.Fatal("envelope not detected")
	}
	if got := env.Error.Error(); !strings.Contains(got, "--nope") {
		t.Errorf("lost the CLI reason: %q", got)
	}
}
