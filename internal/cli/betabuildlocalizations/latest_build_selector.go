package betabuildlocalizations

import (
	"context"
	"fmt"
	"strings"

	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/asc"
	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/cli/shared"
)

func resolveLatestBuildIDForBetaBuildLocalizations(
	ctx context.Context,
	client *asc.Client,
	appInput string,
	stateFilter string,
) (string, error) {
	resolvedAppID := shared.ResolveAppID(appInput)
	if resolvedAppID == "" {
		return "", shared.UsageError("--app is required with --latest")
	}

	resolvedAppID, err := shared.ResolveAppIDWithLookup(ctx, client, resolvedAppID)
	if err != nil {
		return "", err
	}

	stateValues, err := normalizeLatestBuildProcessingStateFilter(stateFilter)
	if err != nil {
		return "", err
	}

	opts := []asc.BuildsOption{
		asc.WithBuildsSort("-uploadedDate"),
		asc.WithBuildsLimit(1),
	}
	if len(stateValues) > 0 {
		opts = append(opts, asc.WithBuildsProcessingStates(stateValues))
	}

	builds, err := client.GetBuilds(ctx, resolvedAppID, opts...)
	if err != nil {
		return "", fmt.Errorf("failed to fetch latest build: %w", err)
	}
	if len(builds.Data) == 0 {
		if len(stateValues) > 0 {
			return "", fmt.Errorf("no builds found for app %s matching state filter", resolvedAppID)
		}
		return "", fmt.Errorf("no builds found for app %s", resolvedAppID)
	}

	buildID := strings.TrimSpace(builds.Data[0].ID)
	if buildID == "" {
		return "", fmt.Errorf("latest build is missing an ID for app %s", resolvedAppID)
	}

	return buildID, nil
}

func normalizeLatestBuildProcessingStateFilter(raw string) ([]string, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}

	values := shared.SplitCSVUpper(raw)
	if len(values) == 0 {
		return nil, shared.UsageError("--state must include at least one state")
	}

	if len(values) == 1 && values[0] == "ALL" {
		return []string{
			asc.BuildProcessingStateProcessing,
			asc.BuildProcessingStateFailed,
			asc.BuildProcessingStateInvalid,
			asc.BuildProcessingStateValid,
		}, nil
	}

	allowed := map[string]struct{}{
		asc.BuildProcessingStateValid:      {},
		asc.BuildProcessingStateProcessing: {},
		asc.BuildProcessingStateFailed:     {},
		asc.BuildProcessingStateInvalid:    {},
		"COMPLETE":                         {}, // compatibility alias for VALID
	}

	resolved := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if value == "ALL" {
			return nil, shared.UsageError("--state value \"all\" cannot be combined with other states")
		}
		if _, ok := allowed[value]; !ok {
			return nil, shared.UsageError("--state must be one of PROCESSING, FAILED, INVALID, VALID, COMPLETE, or all")
		}

		if value == "COMPLETE" {
			value = asc.BuildProcessingStateValid
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		resolved = append(resolved, value)
	}

	return resolved, nil
}
