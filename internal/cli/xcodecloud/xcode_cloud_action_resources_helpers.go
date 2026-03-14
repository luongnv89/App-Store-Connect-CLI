package xcodecloud

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/asc"
	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/cli/shared"
)

type xcodeCloudActionResourceGroupConfig struct {
	Name       string
	ShortUsage string
	ShortHelp  string
	LongHelp   string
	List       *ffcli.Command
	Get        *ffcli.Command
}

func newXcodeCloudActionResourceGroupCommand(config xcodeCloudActionResourceGroupConfig) *ffcli.Command {
	fs := flag.NewFlagSet(config.Name, flag.ExitOnError)

	return &ffcli.Command{
		Name:       config.Name,
		ShortUsage: config.ShortUsage,
		ShortHelp:  config.ShortHelp,
		LongHelp:   config.LongHelp,
		FlagSet:    fs,
		UsageFunc:  shared.DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			config.List,
			config.Get,
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

type xcodeCloudActionResourceListConfig struct {
	ShortUsage  string
	ShortHelp   string
	LongHelp    string
	ActionUsage string
	RunUsage    string
	ErrorPrefix string
	FetchPage   func(context.Context, *asc.Client, string, int, string) (asc.PaginatedResponse, error)
}

func newXcodeCloudActionResourceListCommand(config xcodeCloudActionResourceListConfig) *ffcli.Command {
	fs := flag.NewFlagSet("list", flag.ExitOnError)

	actionUsage := strings.TrimSpace(config.ActionUsage)
	if actionUsage == "" {
		actionUsage = "Build action ID to list resources for"
	}
	runUsage := strings.TrimSpace(config.RunUsage)
	if runUsage == "" {
		runUsage = "Build run ID to resolve a single action from"
	}

	actionID := fs.String("action-id", "", actionUsage)
	runID := fs.String("run-id", "", runUsage)
	limit := fs.Int("limit", 0, "Maximum results per page (1-200)")
	next := fs.String("next", "", "Fetch next page using a links.next URL")
	paginate := fs.Bool("paginate", false, "Automatically fetch all pages (aggregate results)")
	output := shared.BindOutputFlags(fs)

	return &ffcli.Command{
		Name:       "list",
		ShortUsage: config.ShortUsage,
		ShortHelp:  config.ShortHelp,
		LongHelp:   config.LongHelp,
		FlagSet:    fs,
		UsageFunc:  shared.DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			return runXcodeCloudActionResourceList(
				ctx,
				*actionID,
				*runID,
				*limit,
				*next,
				*paginate,
				*output.Output,
				*output.Pretty,
				config.ErrorPrefix,
				config.FetchPage,
			)
		},
	}
}

type xcodeCloudActionResourceGetConfig struct {
	ShortUsage  string
	ShortHelp   string
	LongHelp    string
	IDUsage     string
	ErrorPrefix string
	Fetch       func(context.Context, *asc.Client, string) (any, error)
}

func newXcodeCloudActionResourceGetCommand(config xcodeCloudActionResourceGetConfig) *ffcli.Command {
	return shared.BuildIDGetCommand(shared.IDGetCommandConfig{
		FlagSetName: "get",
		Name:        "get",
		ShortUsage:  config.ShortUsage,
		ShortHelp:   config.ShortHelp,
		LongHelp:    config.LongHelp,
		IDFlag:      "id",
		IDUsage:     config.IDUsage,
		ErrorPrefix: config.ErrorPrefix,
		ContextTimeout: func(ctx context.Context) (context.Context, context.CancelFunc) {
			return contextWithXcodeCloudTimeout(ctx, 0)
		},
		Fetch: config.Fetch,
	})
}

type xcodeCloudActionResourceCommandConfig struct {
	Name       string
	ShortUsage string
	ShortHelp  string
	LongHelp   string

	ListShortUsage  string
	ListShortHelp   string
	ListLongHelp    string
	ListActionUsage string
	ListRunUsage    string
	ListErrorPrefix string
	ListFetchPage   func(context.Context, *asc.Client, string, int, string) (asc.PaginatedResponse, error)

	GetShortUsage  string
	GetShortHelp   string
	GetLongHelp    string
	GetIDUsage     string
	GetErrorPrefix string
	GetFetch       func(context.Context, *asc.Client, string) (any, error)
}

func newXcodeCloudActionResourceCommand(config xcodeCloudActionResourceCommandConfig) *ffcli.Command {
	return newXcodeCloudActionResourceGroupCommand(xcodeCloudActionResourceGroupConfig{
		Name:       config.Name,
		ShortUsage: config.ShortUsage,
		ShortHelp:  config.ShortHelp,
		LongHelp:   config.LongHelp,
		List: newXcodeCloudActionResourceListCommand(xcodeCloudActionResourceListConfig{
			ShortUsage:  config.ListShortUsage,
			ShortHelp:   config.ListShortHelp,
			LongHelp:    config.ListLongHelp,
			ActionUsage: config.ListActionUsage,
			RunUsage:    config.ListRunUsage,
			ErrorPrefix: config.ListErrorPrefix,
			FetchPage:   config.ListFetchPage,
		}),
		Get: newXcodeCloudActionResourceGetCommand(xcodeCloudActionResourceGetConfig{
			ShortUsage:  config.GetShortUsage,
			ShortHelp:   config.GetShortHelp,
			LongHelp:    config.GetLongHelp,
			IDUsage:     config.GetIDUsage,
			ErrorPrefix: config.GetErrorPrefix,
			Fetch:       config.GetFetch,
		}),
	})
}

var xcodeCloudIssuesCommandConfig = xcodeCloudActionResourceCommandConfig{
	Name:       "issues",
	ShortUsage: "asc xcode-cloud issues <subcommand> [flags]",
	ShortHelp:  "List Xcode Cloud build issues.",
	LongHelp: `List Xcode Cloud build issues.

Examples:
  asc xcode-cloud issues list --action-id "ACTION_ID"
  asc xcode-cloud issues list --run-id "BUILD_RUN_ID"
  asc xcode-cloud issues get --id "ISSUE_ID"`,
	ListShortUsage: "asc xcode-cloud issues list [flags]",
	ListShortHelp:  "List issues for a build action.",
	ListLongHelp: `List issues for a build action.

Examples:
  asc xcode-cloud issues list --action-id "ACTION_ID"
  asc xcode-cloud issues list --run-id "BUILD_RUN_ID"
  asc xcode-cloud issues list --action-id "ACTION_ID" --output table
  asc xcode-cloud issues list --action-id "ACTION_ID" --limit 50
  asc xcode-cloud issues list --action-id "ACTION_ID" --paginate`,
	ListActionUsage: "Build action ID to list issues for",
	ListRunUsage:    "Build run ID to resolve a single issue action from",
	ListErrorPrefix: "xcode-cloud issues list",
	ListFetchPage: func(ctx context.Context, client *asc.Client, actionID string, limit int, next string) (asc.PaginatedResponse, error) {
		return client.GetCiBuildActionIssues(ctx, actionID,
			asc.WithCiIssuesLimit(limit),
			asc.WithCiIssuesNextURL(next),
		)
	},
	GetShortUsage: "asc xcode-cloud issues get --id \"ISSUE_ID\"",
	GetShortHelp:  "Get details for a build issue.",
	GetLongHelp: `Get details for a build issue.

Examples:
  asc xcode-cloud issues get --id "ISSUE_ID"
  asc xcode-cloud issues get --id "ISSUE_ID" --output table`,
	GetIDUsage:     "Issue ID",
	GetErrorPrefix: "xcode-cloud issues get",
	GetFetch: func(ctx context.Context, client *asc.Client, id string) (any, error) {
		return client.GetCiIssue(ctx, id)
	},
}

var xcodeCloudTestResultsCommandConfig = xcodeCloudActionResourceCommandConfig{
	Name:       "test-results",
	ShortUsage: "asc xcode-cloud test-results <subcommand> [flags]",
	ShortHelp:  "List Xcode Cloud test results.",
	LongHelp: `List Xcode Cloud test results.

Examples:
  asc xcode-cloud test-results list --action-id "ACTION_ID"
  asc xcode-cloud test-results get --id "TEST_RESULT_ID"`,
	ListShortUsage: "asc xcode-cloud test-results list [flags]",
	ListShortHelp:  "List test results for a build action.",
	ListLongHelp: `List test results for a build action.

Examples:
  asc xcode-cloud test-results list --action-id "ACTION_ID"
  asc xcode-cloud test-results list --action-id "ACTION_ID" --output table
  asc xcode-cloud test-results list --action-id "ACTION_ID" --limit 50
  asc xcode-cloud test-results list --action-id "ACTION_ID" --paginate`,
	ListActionUsage: "Build action ID to list test results for",
	ListRunUsage:    "Build run ID to resolve a single test-result action from",
	ListErrorPrefix: "xcode-cloud test-results list",
	ListFetchPage: func(ctx context.Context, client *asc.Client, actionID string, limit int, next string) (asc.PaginatedResponse, error) {
		return client.GetCiBuildActionTestResults(ctx, actionID,
			asc.WithCiTestResultsLimit(limit),
			asc.WithCiTestResultsNextURL(next),
		)
	},
	GetShortUsage: "asc xcode-cloud test-results get --id \"TEST_RESULT_ID\"",
	GetShortHelp:  "Get details for a test result.",
	GetLongHelp: `Get details for a test result.

Examples:
  asc xcode-cloud test-results get --id "TEST_RESULT_ID"
  asc xcode-cloud test-results get --id "TEST_RESULT_ID" --output table`,
	GetIDUsage:     "Test result ID",
	GetErrorPrefix: "xcode-cloud test-results get",
	GetFetch: func(ctx context.Context, client *asc.Client, id string) (any, error) {
		return client.GetCiTestResult(ctx, id)
	},
}

func runXcodeCloudActionResourceList(
	ctx context.Context,
	actionID string,
	runID string,
	limit int,
	next string,
	paginate bool,
	output string,
	pretty bool,
	errorPrefix string,
	fetchPage func(context.Context, *asc.Client, string, int, string) (asc.PaginatedResponse, error),
) error {
	if strings.TrimSpace(actionID) != "" && strings.TrimSpace(runID) != "" {
		return shared.UsageError("--action-id and --run-id are mutually exclusive")
	}
	if limit != 0 && (limit < 1 || limit > 200) {
		return fmt.Errorf("%s: --limit must be between 1 and 200", errorPrefix)
	}

	nextURL := strings.TrimSpace(next)
	if err := shared.ValidateNextURL(nextURL); err != nil {
		return fmt.Errorf("%s: %w", errorPrefix, err)
	}

	resolvedActionID := strings.TrimSpace(actionID)
	resolvedRunID := strings.TrimSpace(runID)
	if resolvedActionID == "" && resolvedRunID == "" && nextURL == "" {
		return shared.UsageError("--action-id or --run-id is required")
	}

	client, err := shared.GetASCClient()
	if err != nil {
		return fmt.Errorf("%s: %w", errorPrefix, err)
	}

	requestCtx, cancel := contextWithXcodeCloudTimeout(ctx, 0)
	defer cancel()

	if nextURL == "" && resolvedActionID == "" {
		resolvedActionID, err = resolveSingleBuildActionIDForRun(requestCtx, client, resolvedRunID)
		if err != nil {
			if errors.Is(err, flag.ErrHelp) {
				return err
			}
			return fmt.Errorf("%s: %w", errorPrefix, err)
		}
	}

	if paginate {
		resp, err := shared.PaginateWithSpinner(requestCtx,
			func(ctx context.Context) (asc.PaginatedResponse, error) {
				return fetchPage(ctx, client, resolvedActionID, 200, nextURL)
			},
			func(ctx context.Context, nextURL string) (asc.PaginatedResponse, error) {
				return fetchPage(ctx, client, resolvedActionID, 0, nextURL)
			},
		)
		if err != nil {
			return fmt.Errorf("%s: %w", errorPrefix, err)
		}
		return shared.PrintOutput(resp, output, pretty)
	}

	resp, err := fetchPage(requestCtx, client, resolvedActionID, limit, nextURL)
	if err != nil {
		return fmt.Errorf("%s: %w", errorPrefix, err)
	}

	return shared.PrintOutput(resp, output, pretty)
}

func resolveSingleBuildActionIDForRun(ctx context.Context, client *asc.Client, runID string) (string, error) {
	resp, err := client.GetCiBuildActions(ctx, runID, asc.WithCiBuildActionsLimit(2))
	if err != nil {
		return "", fmt.Errorf("resolve build actions for run %q: %w", runID, err)
	}

	if len(resp.Data) == 0 {
		return "", shared.UsageErrorf("no build actions found for --run-id %q", runID)
	}
	if len(resp.Data) > 1 || strings.TrimSpace(resp.Links.Next) != "" {
		return "", shared.UsageErrorf("--run-id %q matched multiple build actions; use --action-id or inspect asc xcode-cloud actions --run-id %q", runID, runID)
	}

	return strings.TrimSpace(resp.Data[0].ID), nil
}
