package metadata

import (
	"flag"

	"github.com/peterbourgon/ff/v3/ffcli"
)

// MetadataApplyCommand returns the canonical apply alias for metadata push.
func MetadataApplyCommand() *ffcli.Command {
	cmd := MetadataPushCommand()
	cmd.Name = "apply"
	cmd.FlagSet = cloneFlagSetWithName(cmd.FlagSet, "metadata apply")
	cmd.ShortUsage = "asc metadata apply --app \"APP_ID\" --version \"1.2.3\" --dir \"./metadata\" [--app-info \"APP_INFO_ID\"] [--dry-run]"
	cmd.ShortHelp = "Apply metadata changes from canonical files."
	cmd.LongHelp = `Apply metadata changes from canonical files.

Examples:
  asc metadata apply --app "APP_ID" --version "1.2.3" --dir "./metadata" --dry-run
  asc metadata apply --app "APP_ID" --version "1.2.3" --platform IOS --dir "./metadata" --dry-run
  asc metadata apply --app "APP_ID" --app-info "APP_INFO_ID" --version "1.2.3" --platform IOS --dir "./metadata" --dry-run
  asc metadata apply --app "APP_ID" --version "1.2.3" --dir "./metadata"
  asc metadata apply --app "APP_ID" --version "1.2.3" --dir "./metadata" --allow-deletes --confirm

Notes:
  - default.json fallback is applied only when --allow-deletes is not set.
  - with --allow-deletes, remote locales missing locally are planned as deletes.
  - omitted fields are treated as no-op; they do not imply deletion.`
	return cmd
}

func cloneFlagSetWithName(src *flag.FlagSet, name string) *flag.FlagSet {
	if src == nil {
		return flag.NewFlagSet(name, flag.ExitOnError)
	}

	dst := flag.NewFlagSet(name, flag.ExitOnError)
	src.VisitAll(func(f *flag.Flag) {
		dst.Var(f.Value, f.Name, f.Usage)
		if cloned := dst.Lookup(f.Name); cloned != nil {
			cloned.DefValue = f.DefValue
		}
	})
	return dst
}
