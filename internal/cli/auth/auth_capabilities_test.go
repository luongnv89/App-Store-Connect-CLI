package auth

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"strings"
	"testing"

	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/cli/shared"
)

func TestAuthCapabilitiesCommandFlagValidation(t *testing.T) {
	t.Run("unsupported output", func(t *testing.T) {
		cmd := AuthCapabilitiesCommand()
		if err := cmd.FlagSet.Parse([]string{"--output", "yaml"}); err != nil {
			t.Fatalf("Parse() error: %v", err)
		}

		_, stderr := captureAuthOutput(t, func() {
			err := cmd.Exec(context.Background(), []string{})
			if !errors.Is(err, flag.ErrHelp) {
				t.Fatalf("expected flag.ErrHelp, got %v", err)
			}
		})
		if !strings.Contains(stderr, "unsupported format") {
			t.Fatalf("expected unsupported format error, got %q", stderr)
		}
	})

	t.Run("pretty requires json output", func(t *testing.T) {
		cmd := AuthCapabilitiesCommand()
		if err := cmd.FlagSet.Parse([]string{"--output", "table", "--pretty"}); err != nil {
			t.Fatalf("Parse() error: %v", err)
		}

		_, stderr := captureAuthOutput(t, func() {
			err := cmd.Exec(context.Background(), []string{})
			if !errors.Is(err, flag.ErrHelp) {
				t.Fatalf("expected flag.ErrHelp, got %v", err)
			}
		})
		if !strings.Contains(stderr, "--pretty is only valid with JSON output") {
			t.Fatalf("expected pretty/json error, got %q", stderr)
		}
	})
}

func TestDefaultAuthCapabilitiesOutputFormat(t *testing.T) {
	t.Run("json", func(t *testing.T) {
		t.Setenv("ASC_DEFAULT_OUTPUT", "json")
		shared.ResetDefaultOutputFormat()
		t.Cleanup(shared.ResetDefaultOutputFormat)

		if got := defaultAuthCapabilitiesOutputFormat(); got != "json" {
			t.Fatalf("defaultAuthCapabilitiesOutputFormat() = %q, want %q", got, "json")
		}
	})

	t.Run("markdown", func(t *testing.T) {
		t.Setenv("ASC_DEFAULT_OUTPUT", "markdown")
		shared.ResetDefaultOutputFormat()
		t.Cleanup(shared.ResetDefaultOutputFormat)

		if got := defaultAuthCapabilitiesOutputFormat(); got != "markdown" {
			t.Fatalf("defaultAuthCapabilitiesOutputFormat() = %q, want %q", got, "markdown")
		}
	})
}

func TestSummarizeAuthCapabilities(t *testing.T) {
	red := summarizeAuthCapabilities([]authCapabilityCheck{
		{Name: "apps", Status: "available"},
		{Name: "analytics", Status: "inconclusive", Message: "analytics probe failed"},
	})
	if red.Health != "red" || red.InconclusiveCount != 1 || red.NextAction != "analytics probe failed" {
		t.Fatalf("unexpected red summary: %+v", red)
	}

	yellow := summarizeAuthCapabilities([]authCapabilityCheck{
		{Name: "apps", Status: "available"},
		{Name: "sales", Status: "unavailable", Message: "sales unavailable"},
	})
	if yellow.Health != "yellow" || yellow.UnavailableCount != 1 || yellow.NextAction != "sales unavailable" {
		t.Fatalf("unexpected yellow summary: %+v", yellow)
	}

	green := summarizeAuthCapabilities([]authCapabilityCheck{
		{Name: "apps", Status: "available"},
		{Name: "builds", Status: "skipped", Message: "provide --app"},
	})
	if green.Health != "green" || green.SkippedCount != 1 || green.NextAction != "provide --app" {
		t.Fatalf("unexpected green summary: %+v", green)
	}
}

func TestAuthCapabilitiesCommandJSONOutput(t *testing.T) {
	prevCollector := authCapabilitiesCollector
	authCapabilitiesCollector = func(context.Context, string, string) (*authCapabilitiesResponse, error) {
		return &authCapabilitiesResponse{
			Summary: authCapabilitiesSummary{
				Health:         "green",
				NextAction:     "No action needed.",
				AvailableCount: 1,
			},
			Inputs: authCapabilitiesInputs{
				AppID:        "123456789",
				VendorNumber: "98765432",
			},
			Capabilities: []authCapabilityCheck{
				{Name: "apps", Scope: "account", Status: "available", Message: "can list apps"},
			},
			GeneratedAt: "2026-03-12T00:00:00Z",
		}, nil
	}
	t.Cleanup(func() {
		authCapabilitiesCollector = prevCollector
	})

	cmd := AuthCapabilitiesCommand()
	if err := cmd.FlagSet.Parse([]string{"--output", "json"}); err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	stdout, _ := captureAuthOutput(t, func() {
		if err := cmd.Exec(context.Background(), []string{}); err != nil {
			t.Fatalf("Exec() error: %v", err)
		}
	})

	var payload authCapabilitiesResponse
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}
	if payload.Summary.Health != "green" || len(payload.Capabilities) != 1 {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}
