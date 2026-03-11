package builds

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/cli/shared"
)

func buildsVisibleUsageFunc(c *ffcli.Command) string {
	clone := *c
	if len(c.Subcommands) > 0 {
		visible := make([]*ffcli.Command, 0, len(c.Subcommands))
		for _, sub := range c.Subcommands {
			if sub == nil {
				continue
			}
			if strings.HasPrefix(strings.TrimSpace(sub.ShortHelp), "DEPRECATED:") {
				continue
			}
			visible = append(visible, sub)
		}
		clone.Subcommands = visible
	}
	return shared.DefaultUsageFunc(&clone)
}

func deprecatedBuildsAliasLeafCommand(cmd *ffcli.Command, name, shortUsage, newCommand, warning string) *ffcli.Command {
	if cmd == nil {
		return nil
	}

	clone := *cmd
	clone.Name = name
	clone.ShortUsage = shortUsage
	clone.ShortHelp = fmt.Sprintf("DEPRECATED: use `%s`.", newCommand)
	clone.LongHelp = fmt.Sprintf("Deprecated compatibility alias for `%s`.", newCommand)
	clone.UsageFunc = shared.DeprecatedUsageFunc

	origExec := cmd.Exec
	clone.Exec = func(ctx context.Context, args []string) error {
		fmt.Fprintln(os.Stderr, warning)
		return origExec(ctx, args)
	}

	return &clone
}

func deprecatedBuildsRelationshipsAliasCommand() *ffcli.Command {
	fs := BuildsRelationshipsCommand().FlagSet

	return &ffcli.Command{
		Name:       "relationships",
		ShortUsage: "asc builds links <subcommand> [flags]",
		ShortHelp:  "DEPRECATED: use `asc builds links ...`.",
		LongHelp:   "Deprecated compatibility alias for `asc builds links ...`.",
		FlagSet:    fs,
		UsageFunc:  shared.DeprecatedUsageFunc,
		Subcommands: []*ffcli.Command{
			deprecatedBuildsAliasLeafCommand(
				BuildsRelationshipsGetCommand(),
				"get",
				"asc builds links view --build \"BUILD_ID\" --type \"RELATIONSHIP\" [flags]",
				"asc builds links view",
				"Warning: `asc builds relationships get` is deprecated. Use `asc builds links view`.",
			),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}
